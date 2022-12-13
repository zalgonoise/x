package spanner

import (
	"time"

	json "github.com/goccy/go-json"
	"github.com/zalgonoise/attr"
)

// var buf = new(bytes.Buffer)

// SpanData is the output data that was recorded by a Span
//
// It contains all the details stored in the Span, and it is returned either
// with the `Extract()` method, or when `End()` is called (and its returned value captured)
type SpanData struct {
	TraceID    TraceID
	SpanID     SpanID
	ParentID   *SpanID
	Name       string
	StartTime  time.Time
	EndTime    *time.Time
	Attributes attr.Attrs
	Events     []EventData
}

// MarshalJSON encodes the SpanData into a byte slice, returning it and an error
func (s SpanData) MarshalJSON() ([]byte, error) {
	// buf.Reset()
	// buf.WriteString(`{"name":"`)
	// buf.WriteString(s.Name)
	// buf.WriteString(`","context":{"trace_id":"`)
	// buf.WriteString(s.TraceID.String())
	// buf.WriteString(`","span_id":"`)
	// buf.WriteString(s.SpanID.String())
	// buf.WriteString(`"},"parent_id":`)
	// if s.ParentID == nil {
	// 	buf.WriteString(`null`)
	// } else {
	// 	buf.WriteString(s.ParentID.String())
	// }
	// buf.WriteString(`,"start_time":"`)
	// buf.WriteString(s.StartTime.Format(time.RFC3339Nano))
	// buf.WriteString(`","end_time":`)
	// if s.ParentID == nil {
	// 	buf.WriteString(`null`)
	// } else {
	// 	buf.WriteString(s.EndTime.Format(time.RFC3339Nano))
	// }
	// if len(s.Attributes) > 0 {
	// 	buf.WriteString(`,"attributes":`)
	// 	attr, _ := json.Marshal(s.Attributes)
	// 	buf.Write(attr)
	// }
	// if len(s.Events) > 0 {
	// 	buf.WriteString(`,"events":`)
	// 	evt, _ := json.Marshal(s.Events)
	// 	buf.Write(evt)
	// }
	// return buf.Bytes(), nil

	type exportedContext struct {
		TraceID TraceID `json:"trace_id"`
		SpanID  SpanID  `json:"span_id"`
	}

	type exportedSpanData struct {
		Name       string          `json:"name"`
		Context    exportedContext `json:"context"`
		ParentID   *SpanID         `json:"parent_id"`
		StartTime  time.Time       `json:"start_time"`
		EndTime    *time.Time      `json:"end_time"`
		Attributes attr.Attrs      `json:"attributes,omitempty"`
		Events     []EventData     `json:"events,omitempty"`
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
		Attributes: s.Attributes,
		Events:     s.Events,
	})
}

// String implements fmt.Stringer
// func (s SpanData) String() string {
// 	b, _ := s.MarshalJSON()
// 	return string(b)
// }

type EventData struct {
	Name       string     `json:"name"`
	Timestamp  time.Time  `json:"timestamp"`
	Attributes attr.Attrs `json:"attributes"`
}

// MarshalJSON encodes the EventData into a byte slice, returning it and an error
func (e EventData) MarshalJSON() ([]byte, error) {
	type exportedEvent struct {
		Name       string     `json:"name"`
		Timestamp  string     `json:"timestamp"`
		Attributes attr.Attrs `json:"attributes,omitempty"`
	}
	return json.Marshal(exportedEvent{
		Name:       e.Name,
		Timestamp:  e.Timestamp.Format(time.RFC3339Nano),
		Attributes: e.Attributes,
	})
}

// String implements fmt.Stringer
func (e EventData) String() string {
	b, _ := e.MarshalJSON()
	return string(b)
}
