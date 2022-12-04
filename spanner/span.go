package spanner

import (
	"time"

	"github.com/zalgonoise/logx/attr"
)

// Span is a single action within a Trace, which holds metadata about
// the action's execution, as well as optional attributes
type Span interface {
	// Start sets the span to record
	Start()
	// End stops the span, returning the collected SpanData in the action
	End() SpanData
	// Extract returns the current SpanData for the Span, regardless of its status
	Extract() SpanData
	// ID returns the SpanID of the Span
	ID() SpanID
	// IsRecording returns a boolean on whether the Span is currently recording
	IsRecording() bool
	// SetName overwrites the Span's name field with the string `name`
	SetName(name string)
	// SetParent overwrites the Span's parent_id field with the SpanID `id`
	SetParent(id *SpanID)
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
	rec     bool
	traceID TraceID
	spanID  SpanID
	parent  *SpanID
	rcv     chan Span

	name   string
	start  time.Time
	end    time.Time
	data   []attr.Attr
	events []event
}

func newSpan(rcv chan Span, tid TraceID, pid *SpanID, name string, attrs ...attr.Attr) Span {
	return &span{
		traceID: tid,
		spanID:  NewSpanID(),
		rcv:     rcv,
		parent:  pid,

		name: name,
		data: attrs,
	}
}

// ID returns the SpanID of the Span
func (s *span) ID() SpanID {
	return s.spanID
}

// Start sets the span to record
func (s *span) Start() {
	s.start = time.Now()
	s.rec = true
}

// End stops the span, returning the collected SpanData in the action
func (s *span) End() SpanData {
	s.end = time.Now()
	s.rec = false

	s.rcv <- s
	return s.Extract()
}

// Add appends attributes (key-value pairs) to the Span
func (s *span) Add(attrs ...attr.Attr) {
	s.data = append(s.data, attrs...)
}

// IsRecording returns a boolean on whether the Span is currently recording
func (s *span) IsRecording() bool {
	return s.rec
}

// SetName overwrites the Span's name field with the string `name`
func (s *span) SetName(name string) {
	s.name = name
}

// SetParent overwrites the Span's parent_id field with the SpanID `id`
func (s *span) SetParent(id *SpanID) {
	if id != nil && id.IsValid() {
		s.parent = id
	}
	s.parent = nil
}

// Attrs returns the Span's stored attributes
func (s *span) Attrs() []attr.Attr {
	return s.data
}

// Replace will flush the Span's attributes and store the input attributes `attrs` in place
func (s *span) Replace(attrs ...attr.Attr) {
	s.data = attrs
}

// Extract returns the current SpanData for the Span, regardless of its status
func (s *span) Extract() SpanData {
	var parentID *string = nil
	if s.parent != nil {
		pid := s.parent.String()
		parentID = &pid
	}
	var endTime = s.end.Format(time.RFC3339Nano)

	return SpanData{
		TraceID:    s.traceID.String(),
		SpanID:     s.spanID.String(),
		ParentID:   parentID,
		Name:       s.name,
		StartTime:  s.start.Format(time.RFC3339Nano),
		EndTime:    &endTime,
		Attributes: s.data,
		Events:     s.Events(),
	}
}

// Event creates a new event within the Span
func (s *span) Event(name string, attrs ...attr.Attr) {
	s.events = append(s.events, newEvent(name, attrs...))
}

// Events returns the events in the Span
func (s *span) Events() []EventData {
	eventData := make([]EventData, len(s.events))
	for idx, e := range s.events {
		eventData[idx] = EventData{
			Name:       e.name,
			Timestamp:  e.timestamp,
			Attributes: e.attrs,
		}
	}
	return eventData
}
