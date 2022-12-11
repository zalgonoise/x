package spanner

import (
	"time"

	json "github.com/goccy/go-json"
	"github.com/zalgonoise/attr"
)

// SpanData is the output data that was recorded by a Span
//
// It contains all the details stored in the Span, and it is returned either
// with the `Extract()` method, or when `End()` is called (and its returned value captured)
type SpanData struct {
	TraceID    TraceID     `json:"trace_id"`
	SpanID     SpanID      `json:"span_id"`
	ParentID   *SpanID     `json:"parent_id"`
	Name       string      `json:"name"`
	StartTime  time.Time   `json:"start_time"`
	EndTime    *time.Time  `json:"end_time"`
	Attributes attr.Attrs  `json:"attributes,omitempty"`
	Events     []EventData `json:"events,omitempty"`
}

// MarshalJSON encodes the SpanData into a byte slice, returning it and an error
func (s SpanData) MarshalJSON() ([]byte, error) {
	type exportedContext struct {
		TraceID string `json:"trace_id"`
		SpanID  string `json:"span_id"`
	}

	type exportedSpanData struct {
		Name       string          `json:"name"`
		Context    exportedContext `json:"context"`
		ParentID   *string         `json:"parent_id"`
		StartTime  string          `json:"start_time"`
		EndTime    string          `json:"end_time"`
		Attributes map[string]any  `json:"attributes,omitempty"`
		Events     []EventData     `json:"events,omitempty"`
	}

	var parentID *string = nil
	if s.ParentID != nil {
		pid := s.ParentID.String()
		parentID = &pid
	}
	var endTime = "<nil>"
	if s.EndTime != nil {
		endTime = s.EndTime.Format(time.RFC3339Nano)
	}

	return json.Marshal(exportedSpanData{
		Name: s.Name,
		Context: exportedContext{
			TraceID: s.TraceID.String(),
			SpanID:  s.SpanID.String(),
		},
		ParentID:   parentID,
		StartTime:  s.StartTime.Format(time.RFC3339Nano),
		EndTime:    endTime,
		Attributes: attr.Map(s.Attributes...),
		Events:     s.Events,
	})
}

// String implements fmt.Stringer
func (s SpanData) String() string {
	b, _ := s.MarshalJSON()
	return string(b)
}

type EventData struct {
	Name       string     `json:"name"`
	Timestamp  time.Time  `json:"timestamp"`
	Attributes attr.Attrs `json:"attributes"`
}

// MarshalJSON encodes the EventData into a byte slice, returning it and an error
func (e EventData) MarshalJSON() ([]byte, error) {
	type exportedEvent struct {
		Name       string         `json:"name"`
		Timestamp  string         `json:"timestamp"`
		Attributes map[string]any `json:"attributes,omitempty"`
	}
	return json.Marshal(exportedEvent{
		Name:       e.Name,
		Timestamp:  e.Timestamp.Format(time.RFC3339Nano),
		Attributes: attr.Map(e.Attributes...),
	})
}

// String implements fmt.Stringer
func (e EventData) String() string {
	b, _ := e.MarshalJSON()
	return string(b)
}
