package spanner

import (
	"time"

	"github.com/zalgonoise/logx/attr"
)

// Span is a single action within a Trace, which holds metadata about
// the action's execution, as well as optional attributes
type Span interface {
	// ID returns the SpanID of the Span
	ID() SpanID
	// Start sets the span to record
	Start()
	// End stops the span, returning the collected SpanData in the action
	End() SpanData
	// Add appends attributes (key-value pairs) to the Span
	Add(attrs ...attr.Attr)
	// IsRecording returns a boolean on whether the Span is currently recording
	IsRecording() bool
	// SetName overwrites the Span's name field with the string `name`
	SetName(name string)
	// Attrs returns the Span's stored attributes
	Attrs() []attr.Attr
	// Replace will flush the Span's attributes and store the input attributes `attrs` in place
	Replace(attrs ...attr.Attr)
	// Extract returns the current SpanData for the Span, regardless of its status
	Extract() SpanData
}

// SpanData is the output data that was recorded by a Span
//
// It contains all the details stored in the Span, and it is returned either
// with the `Extract()` method, or when `End()` is called (and its returned value captured)
type SpanData struct {
	TraceID    string      `json:"trace_id"`
	SpanID     string      `json:"span_id"`
	ParentID   *string     `json:"parent_id"`
	Name       string      `json:"name"`
	StartTime  string      `json:"start_time"`
	EndTime    *string     `json:"end_time"`
	Attributes []attr.Attr `json:"attributes"`
}

type span struct {
	rec     bool
	traceID TraceID
	spanID  SpanID
	parent  Trace

	name  string
	start time.Time
	end   time.Time
	data  []attr.Attr
}

func newSpan(t Trace, name string, attrs ...attr.Attr) Span {
	return &span{
		traceID: t.ID(),
		spanID:  NewSpanID(),
		parent:  t,

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

	var parentID *string = nil
	if s.parent.PID() != nil {
		pID := s.parent.PID().String()
		parentID = &pID
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
	}
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
	if s.parent.PID() != nil {
		pID := s.parent.PID().String()
		parentID = &pID
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
	}
}

// AsAttr returns the SpanData as an Attr
func (s SpanData) AsAttr() attr.Attr {
	return attr.New(
		"span", []attr.Attr{
			attr.String("name", s.Name),
			attr.New("context", []attr.Attr{
				attr.String("trace_id", s.TraceID),
				attr.String("span_id", s.SpanID),
			}),
			attr.Ptr("parent_id", s.ParentID),
			attr.String("start_time", s.StartTime),
			attr.Ptr("end_time", s.EndTime),
			attr.New("attributes", s.Attributes),
		})
}
