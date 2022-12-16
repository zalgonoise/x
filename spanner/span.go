package spanner

import (
	"sync"
	"time"

	"github.com/zalgonoise/attr"
)

// Span is a single action within a Trace, which holds metadata about
// the action's execution, as well as optional attributes and events
type Span interface {
	// Start sets the span to record
	Start()
	// End stops the span, returning the collected SpanData in the action
	End()
	// ID returns the SpanID of the Span
	ID() SpanID
	// IsRecording returns a boolean on whether the Span is currently recording
	IsRecording() bool
	// SetName overwrites the Span's name field with the string `name`
	SetName(name string)
	// SetParent overwrites the Span's parent_id field with the SpanID `id`
	SetParent(span Span)
	// Add appends attributes (key-value pairs) to the Span
	Add(attrs ...attr.Attr)
	// Attrs returns the Span's stored attributes
	Attrs() []attr.Attr
	// Replace will flush the Span's attributes and store the input attributes `attrs` in place
	Replace(attrs ...attr.Attr)
	// Event creates a new event within the Span
	Event(name string, attrs ...attr.Attr)
	// Extract returns the current SpanData for the Span, regardless of its status
	Extract() SpanData
	// Events returns the events in the Span
	Events() []EventData
}

type span struct {
	rec bool
	sync.RWMutex
	trace  Trace
	spanID SpanID
	parent *SpanID

	name   string
	start  time.Time
	end    *time.Time
	attrs  []attr.Attr
	events []*event
}

func newSpan(trace Trace, name string) Span {
	newSpan := &span{
		trace:  trace,
		spanID: NewSpanID(),
		parent: trace.Parent(),
		name:   name,
	}
	return newSpan
}

// Start sets the span to record
func (s *span) Start() {
	if s.end != nil {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.start = time.Now()
	s.rec = true
}

// End stops the span, returning the collected SpanData in the action
func (s *span) End() {
	if s.end != nil {
		return
	}
	t := time.Now()
	s.end = new(time.Time)

	s.Lock()
	*s.end = t
	s.rec = false
	s.Unlock()

	p := proc.Load().(SpanProcessor)
	p.Handle(s)
}

// ID returns the SpanID of the Span
func (s *span) ID() SpanID {
	return s.spanID
}

// IsRecording returns a boolean on whether the Span is currently recording
func (s *span) IsRecording() bool {
	return s.rec
}

// SetName overwrites the Span's name field with the string `name`
func (s *span) SetName(name string) {
	if s.end != nil {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.name = name
}

// SetParent overwrites the Span's parent_id field with the SpanID `id`
func (s *span) SetParent(span Span) {
	if s.end != nil {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.parent = nil
	if span != nil {
		sid := span.ID()
		s.parent = &sid
	}
	return
}

// Add appends attributes (key-value pairs) to the Span
func (s *span) Add(attrs ...attr.Attr) {
	if len(attrs) == 0 || s.end != nil {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.attrs = append(s.attrs, attrs...)
}

// Attrs returns the Span's stored attributes
func (s *span) Attrs() []attr.Attr {
	return s.attrs
}

// Replace will flush the Span's attributes and store the input attributes `attrs` in place
func (s *span) Replace(attrs ...attr.Attr) {
	if s.end != nil {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.attrs = attrs
}

// Event creates a new event within the Span
func (s *span) Event(name string, attrs ...attr.Attr) {
	if s.end != nil || name == "" && len(attrs) == 0 {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.events = append(s.events, newEvent(name, attrs...))
}

// Extract returns the current SpanData for the Span, regardless of its status
func (s *span) Extract() SpanData {
	if s.end != nil && s.rec {
		s.Lock()
		defer s.Unlock()
	}

	return SpanData{
		TraceID:    s.trace.ID(),
		SpanID:     s.spanID,
		ParentID:   s.parent,
		Name:       s.name,
		StartTime:  s.start,
		EndTime:    s.end,
		Attributes: s.attrs,
		Events:     s.Events(),
	}
}

// Events returns the events in the Span
func (s *span) Events() []EventData {
	if s.end != nil && s.rec {
		s.Lock()
		defer s.Unlock()
	}

	eventData := make([]EventData, len(s.events), len(s.events))
	for idx, e := range s.events {
		eventData[idx] = EventData{
			Name:       e.name,
			Timestamp:  e.timestamp,
			Attributes: e.attrs,
		}
	}
	return eventData
}
