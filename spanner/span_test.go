package spanner

import (
	"bytes"
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zalgonoise/attr"
)

var (
	testProc = atomic.Value{}
	testBuf  = new(bytes.Buffer)
)

func TestNewSpan(t *testing.T) {
	var (
		name  = "test"
		attrs = []attr.Attr{
			attr.String("attr", "attr"),
			attr.Int("idx", 0),
		}
		tr = newTrace()
		sp = newSpan(tr, name)
	)
	testProc.Store(NewProcessor(Writer(testBuf)))
	defer testProc.Load().(SpanProcessor).Shutdown(context.Background())

	t.Run("Simple", func(t *testing.T) {
		newSpan := newSpan(tr, name)
		s, ok := (newSpan).(*span)
		if !ok {
			t.Errorf("failed to cast Span as *span")
		}

		if s.name != name {
			t.Errorf("unexpected output error: wanted %v ; got %v", name, s.name)
		}
		if s.rec {
			t.Errorf("expected the Span not to be recording since it didn't start yet")
		}
		if !s.trace.ID().IsValid() {
			t.Errorf("invalid TraceID")
		}
		if !s.spanID.IsValid() {
			t.Errorf("invalid SpanID")
		}
		if s.trace.ID().String() != tr.ID().String() {
			t.Errorf("unexpected output error: wanted %s ; got %s", tr.ID().String(), s.trace.ID().String())
		}
		if s.parent != nil {
			t.Error("expected parent's SpanID to be nil")
		}
		if !s.start.IsZero() || !reflect.DeepEqual(time.Time{}, s.start) {
			t.Errorf("invalid start time: %v", s.start)
		}
		if s.end != nil {
			t.Errorf("expected end to be nil: got %v", s.end)
		}
		if len(s.attrs) != 0 {
			t.Errorf("expected empty attribute list")
		}
		if len(s.events) != 0 {
			t.Errorf("expected empty events list")
		}
	})
	t.Run("WithAttrs", func(t *testing.T) {
		newSpan := newSpan(tr, name)
		newSpan.Add(attrs...)
		s, ok := (newSpan).(*span)
		if !ok {
			t.Errorf("failed to cast Span as *span")
		}

		if s.name != name {
			t.Errorf("unexpected output error: wanted %v ; got %v", name, s.name)
		}
		if s.rec {
			t.Errorf("expected the Span not to be recording since it didn't start yet")
		}
		if !s.trace.ID().IsValid() {
			t.Errorf("invalid TraceID")
		}
		if !s.spanID.IsValid() {
			t.Errorf("invalid SpanID")
		}
		if s.trace.ID().String() != tr.ID().String() {
			t.Errorf("unexpected output error: wanted %s ; got %s", tr.ID().String(), s.trace.ID().String())
		}
		if s.parent != nil {
			t.Error("expected parent's SpanID to be nil")
		}
		if !s.start.IsZero() || !reflect.DeepEqual(time.Time{}, s.start) {
			t.Errorf("invalid start time: %v", s.start)
		}
		if s.end != nil {
			t.Errorf("expected end to be nil: got %v", s.end)
		}
		if len(s.attrs) != 2 {
			t.Errorf("expected empty attribute list")
		}
		if len(s.events) != 0 {
			t.Errorf("expected empty events list")
		}
	})
	t.Run("WithParent", func(t *testing.T) {
		sid := sp.ID()
		tr.Register(&sid)
		newSpan := newSpan(tr, name)
		s, ok := (newSpan).(*span)
		if !ok {
			t.Errorf("failed to cast Span as *span")
		}

		if s.name != name {
			t.Errorf("unexpected output error: wanted %v ; got %v", name, s.name)
		}
		if s.rec {
			t.Errorf("expected the Span not to be recording since it didn't start yet")
		}
		if !s.trace.ID().IsValid() {
			t.Errorf("invalid TraceID")
		}
		if !s.spanID.IsValid() {
			t.Errorf("invalid SpanID")
		}
		if s.trace.ID().String() != tr.ID().String() {
			t.Errorf("unexpected output error: wanted %s ; got %s", tr.ID().String(), s.trace.ID().String())
		}
		if s.parent == nil {
			t.Error("expected parent's SpanID to be nil")
		}
		if *s.parent != sp.ID() {
			t.Errorf("unexpected output error: wanted %v ; got %v", sp.ID(), *s.parent)
		}
		if !s.start.IsZero() || !reflect.DeepEqual(time.Time{}, s.start) {
			t.Errorf("invalid start time: %v", s.start)
		}
		if s.end != nil {
			t.Errorf("expected end to be nil: got %v", s.end)
		}
		if len(s.attrs) != 0 {
			t.Errorf("expected empty attribute list")
		}
		if len(s.events) != 0 {
			t.Errorf("expected empty events list")
		}
	})
}

func TestSpanIDMethod(t *testing.T) {
	var (
		name = "test"
		tr   = newTrace()
	)

	testProc.Store(NewProcessor(Writer(testBuf)))
	defer testProc.Load().(SpanProcessor).Shutdown(context.Background())

	testProc.Store(NewProcessor(Writer(testBuf)))
	t.Run("Success", func(t *testing.T) {
		s := newSpan(tr, name)
		id := s.ID()

		if !id.IsValid() {
			t.Errorf("invalid SpanID")
		}
	})
}

func TestSpanStart(t *testing.T) {
	var (
		name = "test"
		tr   = newTrace()
	)
	testProc.Store(NewProcessor(Writer(testBuf)))
	defer testProc.Load().(SpanProcessor).Shutdown(context.Background())

	t.Run("Success", func(t *testing.T) {
		newSpan := newSpan(tr, name)
		s, ok := (newSpan).(*span)
		if !ok {
			t.Errorf("failed to cast Span as *span")
		}

		if s.rec {
			t.Errorf("expected the Span not to be recording since it didn't start yet")
		}
		if !s.start.IsZero() || !reflect.DeepEqual(time.Time{}, s.start) {
			t.Errorf("invalid start time: %v", s.start)
		}

		s.Start()

		if !s.rec {
			t.Errorf("expected the Span not to be recording since it didn't start yet")
		}
		if s.start.IsZero() || reflect.DeepEqual(time.Time{}, s.start) {
			t.Errorf("invalid start time: %v", s.start)
		}
	})
}

func TestSpanAllMethods(t *testing.T) {
	var (
		name    = "test"
		newName = "testing"
		attrs   = []attr.Attr{
			attr.String("attr", "attr"),
			attr.Int("idx", 0),
		}
		tr  = newTrace()
		sp  = newSpan(tr, name)
		pid = sp.ID()
	)

	testProc.Store(NewProcessor(Writer(testBuf)))
	defer testProc.Load().(SpanProcessor).Shutdown(context.Background())
	t.Run("Success", func(t *testing.T) {
		newSpan := newSpan(tr, name)
		s, ok := (newSpan).(*span)
		if !ok {
			t.Errorf("failed to cast Span as *span")
		}

		wants := SpanData{
			Name:       newName,
			TraceID:    tr.ID(),
			ParentID:   &pid,
			SpanID:     s.ID(),
			Attributes: []attr.Attr{attrs[0]},
			Events: []EventData{{
				Name:       name,
				Attributes: []attr.Attr{},
			}},
		}

		if s.name != name {
			t.Errorf("unexpected output error: wanted %v ; got %v", name, s.name)
		}
		if s.rec || s.IsRecording() {
			t.Errorf("expected the Span not to be recording since it didn't start yet")
		}
		if !s.trace.ID().IsValid() {
			t.Errorf("invalid TraceID")
		}
		if !s.spanID.IsValid() {
			t.Errorf("invalid SpanID")
		}
		if s.trace.ID().String() != tr.ID().String() {
			t.Errorf("unexpected output error: wanted %s ; got %s", tr.ID().String(), s.trace.ID().String())
		}
		if s.parent != nil {
			t.Error("expected parent's SpanID to be nil")
		}
		if !s.start.IsZero() || !reflect.DeepEqual(time.Time{}, s.start) {
			t.Errorf("invalid start time: %v", s.start)
		}
		if s.end != nil {
			t.Errorf("expected end to be nil: got %v", s.end)
		}
		if len(s.attrs) != 0 {
			t.Errorf("expected empty attribute list")
		}
		if len(s.events) != 0 {
			t.Errorf("expected empty events list")
		}

		s.Start()

		if !s.rec || !s.IsRecording() {
			t.Errorf("expected the Span to be recording since it started already")
		}
		if s.start.IsZero() || reflect.DeepEqual(time.Time{}, s.start) {
			t.Errorf("invalid start time: %v", s.start)
		}

		s.SetName(newName)
		if s.name != newName {
			t.Errorf("unexpected output error: wanted %v ; got %v", newName, s.name)
		}

		s.SetParent(sp)
		if s.parent == nil {
			t.Error("expected parent's SpanID not to be nil")
			return
		}
		if s.parent.String() != sp.ID().String() {
			t.Errorf("unexpected output errorf: wanted %v ; got %v", sp.ID().String(), s.parent.String())
		}

		s.SetParent(nil)
		if s.parent != nil {
			t.Error("expected parent's SpanID to be nil")
		}

		s.SetParent(sp)
		if s.parent == nil {
			t.Error("expected parent's SpanID not to be nil")
			return
		}
		if s.parent.String() != sp.ID().String() {
			t.Errorf("unexpected output errorf: wanted %v ; got %v", sp.ID().String(), s.parent.String())
		}

		s.Add(attrs...)
		if len(s.attrs) != 2 {
			t.Errorf("expected attribute list with %v element(s)", 2)
		}

		s.Replace(attrs[0])
		if len(s.attrs) != 1 {
			t.Errorf("expected attribute list with %v element(s)", 1)
		}

		a := s.Attrs()
		if len(a) != 1 {
			t.Errorf("expected attribute list with %v element(s)", 1)
		}
		if a[0].Key() != attrs[0].Key() {
			t.Errorf("unexpected output errorf: wanted %v ; got %v", attrs[0].Key(), a[0].Key())
		}
		if a[0].Value().(string) != attrs[0].Value().(string) {
			t.Errorf("unexpected output errorf: wanted %v ; got %v", attrs[0].Value().(string), a[0].Value().(string))
		}

		s.Event(name)
		if len(s.events) != 1 {
			t.Errorf("expected events list with %v element(s)", 1)
		}

		s.End()
		spanData := s.Extract()
		if s.end == nil || s.end.IsZero() || reflect.DeepEqual(time.Time{}, s.end) {
			t.Errorf("expected end to be zero: got %v", s.end)
		}
		if s.rec || s.IsRecording() {
			t.Errorf("expected the Span not to be recording since it already ended")
		}
		spanDataBuf, _ := spanData.MarshalJSON()

		wants.StartTime = s.start
		wants.EndTime = s.end
		wants.Events[0].Timestamp = s.events[0].timestamp

		wantsBuf, _ := wants.MarshalJSON()

		if string(wantsBuf) != string(spanDataBuf) {
			t.Errorf("unexpected output error: wanted %s ; got %s", wantsBuf, spanDataBuf)
		}
	})
}
