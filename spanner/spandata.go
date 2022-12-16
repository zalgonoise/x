package spanner

import (
	"bytes"
	"strings"
	"sync"
	"time"

	json "github.com/goccy/go-json"
	"github.com/zalgonoise/attr"
)

var bufPool = []sync.Pool{
	{New: func() any { return bytes.NewBuffer(make([]byte, 0, 1<<8)) }},
	{New: func() any { return bytes.NewBuffer(make([]byte, 0, 1<<10)) }},
}

// SpanData is the output data that was recorded by a Span
//
// It contains all the details stored in the Span, and it is returned
// with the `Extract()` method
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
	buf := bufPool[1].Get().(*bytes.Buffer)
	defer bufPool[1].Put(buf)
	buf.Reset()
	buf.WriteString(`{"name":"`)

	buf.WriteString(strings.Replace(s.Name, `"`, `\"`, -1))
	buf.WriteString(`","context":{"trace_id":"`)
	buf.WriteString(s.TraceID.String())
	buf.WriteString(`","span_id":"`)
	buf.WriteString(s.SpanID.String())
	buf.WriteString(`"},"parent_id":`)
	if s.ParentID == nil {
		buf.WriteString(`null`)
	} else {
		buf.WriteByte('"')
		buf.WriteString(s.ParentID.String())
		buf.WriteByte('"')
	}
	buf.WriteString(`,"start_time":"`)
	buf.WriteString(s.StartTime.Format(time.RFC3339Nano))
	buf.WriteString(`","end_time":`)
	if s.EndTime == nil {
		buf.WriteString(`null`)
	} else {
		buf.WriteByte('"')
		buf.WriteString(s.EndTime.Format(time.RFC3339Nano))
		buf.WriteByte('"')
	}
	if len(s.Attributes) > 0 {
		buf.WriteString(`,"attributes":`)
		attr, _ := s.Attributes.MarshalJSON()
		buf.Write(attr)
	}
	if len(s.Events) > 0 {
		buf.WriteString(`,"events":[`)
		for idx, evt := range s.Events {
			if idx > 0 {
				buf.WriteByte(',')
			}
			e, _ := evt.MarshalJSON()
			buf.Write(e)
		}
		buf.WriteByte(']')
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// EventData describes the structure of an exported Span Event
type EventData struct {
	Name       string     `json:"name"`
	Timestamp  time.Time  `json:"timestamp"`
	Attributes attr.Attrs `json:"attributes"`
}

// MarshalJSON encodes the EventData into a byte slice, returning it and an error
func (e EventData) MarshalJSON() ([]byte, error) {
	buf := bufPool[0].Get().(*bytes.Buffer)
	defer bufPool[0].Put(buf)
	buf.Reset()
	buf.WriteString(`{"name":"`)
	buf.WriteString(strings.Replace(e.Name, `"`, `\"`, -1))
	buf.WriteString(`","timestamp":"`)
	buf.WriteString(e.Timestamp.Format(time.RFC3339Nano))
	buf.WriteByte('"')
	if len(e.Attributes) > 0 {
		buf.WriteString(`,"attributes":`)
		attr, _ := json.Marshal(e.Attributes)
		buf.Write(attr)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
