package spanner

import (
	"sync"
	"time"

	"github.com/zalgonoise/attr"
)

// Span is a single action within a Trace, which holds metadata about
// the action's execution, as well as optional attributes
type Span interface {
	// Start sets the span to record
	Start()
	// End stops the span, returning the collected SpanData in the action
	End()
	// Extract returns the current SpanData for the Span, regardless of its status
	Extract() SpanData
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

func newSpan(trace Trace, name string, attrs ...attr.Attr) Span {
	var s *SpanID = nil
	if trace.Parent() != nil {
		sid := trace.Parent().ID()
		s = &sid
	}

	return &span{
		trace:  trace,
		spanID: NewSpanID(),
		parent: s,

		name:  name,
		attrs: attrs,
	}
}

// can't overwrite a span
func (s *span) canWrite() bool {
	return s.end == nil
}

// ID returns the SpanID of the Span
func (s *span) ID() SpanID {
	return s.spanID
}

// Start sets the span to record
func (s *span) Start() {
	if !s.canWrite() {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.start = time.Now()
	s.rec = true
}

// End stops the span, returning the collected SpanData in the action
func (s *span) End() {
	if !s.canWrite() {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.end = new(time.Time)
	*s.end = time.Now()
	s.rec = false

	if s.parent == nil {
		defer s.trace.Export()
	}
}

// Add appends attributes (key-value pairs) to the Span
func (s *span) Add(attrs ...attr.Attr) {
	if !s.canWrite() {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.attrs = append(s.attrs, attrs...)
}

// IsRecording returns a boolean on whether the Span is currently recording
func (s *span) IsRecording() bool {
	return s.rec
}

// SetName overwrites the Span's name field with the string `name`
func (s *span) SetName(name string) {

	if !s.canWrite() {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.name = name
}

// SetParent overwrites the Span's parent_id field with the SpanID `id`
func (s *span) SetParent(span Span) {
	if !s.canWrite() {
		return
	}

	s.Lock()
	defer s.Unlock()
	if span != nil {
		sid := span.ID()
		s.parent = &sid
		return
	}
	s.parent = nil
}

// Attrs returns the Span's stored attributes
func (s *span) Attrs() []attr.Attr {
	return s.attrs
}

// Replace will flush the Span's attributes and store the input attributes `attrs` in place
func (s *span) Replace(attrs ...attr.Attr) {
	if !s.canWrite() {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.attrs = attrs
}

// Extract returns the current SpanData for the Span, regardless of its status
func (s *span) Extract() SpanData {
	if s.canWrite() && s.rec {
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

// Event creates a new event within the Span
func (s *span) Event(name string, attrs ...attr.Attr) {
	if !s.canWrite() {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.events = append(s.events, newEvent(name, attrs...))
}

// Events returns the events in the Span
func (s *span) Events() []EventData {
	if s.canWrite() && s.rec {
		s.Lock()
		defer s.Unlock()
	}

	eventData := make([]EventData, 0, len(s.events))
	for _, e := range s.events {
		eventData = append(eventData, EventData{
			Name:       e.name,
			Timestamp:  e.timestamp,
			Attributes: e.attrs,
		})
	}
	return eventData
}
