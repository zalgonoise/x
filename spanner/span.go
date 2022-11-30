package spanner

import (
	"time"

	"github.com/zalgonoise/x/log/attr"
)

type Span interface {
	End() SpanData
	Add(attrs ...attr.Attr)
	IsRecording() bool
	SetName(name string)
	Attrs() []attr.Attr
	Replace(attrs ...attr.Attr)
	Extract() SpanData
}

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
	rec      bool
	traceID  TraceID
	spanID   SpanID
	parentID *SpanID

	name  string
	start time.Time
	end   time.Time
	data  []attr.Attr
}

func (s *span) End() SpanData {
	s.end = time.Now()
	s.rec = false

	var parentID *string = nil
	if s.parentID != nil {
		*parentID = s.parentID.String()
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
func (s *span) Add(attrs ...attr.Attr) {
	s.data = append(s.data, attrs...)
}
func (s *span) IsRecording() bool {
	return s.rec
}
func (s *span) SetName(name string) {
	s.name = name
}
func (s *span) Attrs() []attr.Attr {
	return s.data
}
func (s *span) Replace(attrs ...attr.Attr) {
	s.data = attrs
}
func (s *span) Extract() SpanData {
	var parentID *string = nil
	if s.parentID != nil {
		*parentID = s.parentID.String()
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
