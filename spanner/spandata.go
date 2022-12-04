package spanner

import (
	"time"

	json "github.com/goccy/go-json"
	"github.com/zalgonoise/logx/attr"
)

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
	Attributes []attr.Attr `json:"attributes,omitempty"`
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
		EndTime    *string         `json:"end_time"`
		Attributes map[string]any  `json:"attributes,omitempty"`
		Events     []EventData     `json:"events,omitempty"`
	}

	// TODO: update logx/attr.Attr
	var kv = map[string]any{}
	for _, a := range s.Attributes {
		kv[a.Key()] = a.Value()
	}

	return json.Marshal(exportedSpanData{
		Name: s.Name,
		Context: exportedContext{
			TraceID: s.TraceID,
			SpanID:  s.SpanID,
		},
		ParentID:   s.ParentID,
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
		Attributes: kv,
		Events:     s.Events,
	})
}

// String implements fmt.Stringer
func (s SpanData) String() string {
	b, _ := s.MarshalJSON()
	return string(b)
}

// AsAttr converts a SpanData into an Attr
func AsAttr(s SpanData) attr.Attr {
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

type EventData struct {
	Name       string      `json:"name"`
	Timestamp  time.Time   `json:"timestamp"`
	Attributes []attr.Attr `json:"attributes"`
}

// MarshalJSON encodes the EventData into a byte slice, returning it and an error
func (e EventData) MarshalJSON() ([]byte, error) {
	type exportedEvent struct {
		Name       string         `json:"name"`
		Timestamp  string         `json:"timestamp"`
		Attributes map[string]any `json:"attributes,omitempty"`
	}

	// TODO: update logx/attr.Attr
	var kv = map[string]any{}
	for _, a := range e.Attributes {
		kv[a.Key()] = a.Value()
	}

	return json.Marshal(exportedEvent{
		Name:       e.Name,
		Timestamp:  e.Timestamp.Format(time.RFC3339Nano),
		Attributes: kv,
	})
}

// String implements fmt.Stringer
func (e EventData) String() string {
	b, _ := e.MarshalJSON()
	return string(b)
}
