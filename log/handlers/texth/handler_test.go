package texth

import (
	"bytes"
	"testing"
	"time"

	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/level"
	"github.com/zalgonoise/x/log/records"
)

var (
	testTime  = time.Unix(1668802887, 0)
	testMsg   = "test message"
	testLevel = level.Info

	ta1 = attr.String("a_key", "value")
	ta2 = attr.Int("b_test_no", 1)
	ta3 = attr.Float("c_success_rate", 1.0)
	ta4 = attr.New("d_custom", struct {
		Key   string
		Value int
	}{
		Key:   "custom_key",
		Value: 2,
	})
	ta5    = attr.New("e_slice", []string{"a", "b", "c"})
	taNest = attr.New("namespace", []attr.Attr{
		ta3, ta4,
	})

	testAttrs = []attr.Attr{
		ta1, ta2, ta3, ta4,
	}

	r1 = records.New(testTime, testLevel, testMsg)
	r2 = records.New(testTime, testLevel, testMsg, ta1)
	r3 = records.New(testTime, testLevel, testMsg, testAttrs...)
	r4 = records.New(testTime, testLevel, testMsg, ta1, ta2, taNest)
	r5 = records.New(testTime, testLevel, testMsg, ta5)
)

func TestNew(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := New(b)

		if h == nil {
			t.Errorf("expected output handler not to be nil")
		}
	})
	t.Run("Success", func(t *testing.T) {
		h := New(nil)

		if h != nil {
			t.Errorf("expected output handler to be nil")
		}
	})
}

func TestHandle(t *testing.T) {
	b := &bytes.Buffer{}
	h := New(b)

	t.Run("Simple", func(t *testing.T) {
		b.Reset()
		wants := `[2022-11-18T21:21:27+01:00] [info] test message`

		err := h.Handle(r1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		out := b.String()
		if wants != out {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants, out)
		}
	})
	t.Run("WithAttribute", func(t *testing.T) {
		b.Reset()
		wants := `[2022-11-18T21:21:27+01:00] [info] test message [ a_key: value ]`

		err := h.Handle(r2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		out := b.String()
		if wants != out {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants, out)
		}
	})
	t.Run("WithAttributes", func(t *testing.T) {
		b.Reset()
		wants := `[2022-11-18T21:21:27+01:00] [info] test message [ a_key: value ; b_test_no: 1 ; c_success_rate: 1 ; d_custom: {custom_key 2} ]`

		err := h.Handle(r3)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		out := b.String()
		if wants != out {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants, out)
		}
	})
	t.Run("WithSliceAttribute", func(t *testing.T) {
		b.Reset()
		wants := `[2022-11-18T21:21:27+01:00] [info] test message [ e_slice: [a b c] ]`

		err := h.Handle(r5)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		out := b.String()
		if wants != out {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants, out)
		}
	})
	t.Run("LevelSkip", func(t *testing.T) {
		b.Reset()
		newH := h.WithLevel(level.Warn)

		err := newH.Handle(r3)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		out := b.String()
		if out != "" {
			t.Errorf("output mismatch error: wanted empty string ; got %s", out)
		}
	})
	t.Run("LevelHandlerAttr", func(t *testing.T) {
		wants := `[2022-11-18T21:21:27+01:00] [info] test message [ k: v ; a_key: value ; b_test_no: 1 ; c_success_rate: 1 ; d_custom: {custom_key 2} ]`
		newH := h.With(attr.New("k", "v"))

		err := newH.Handle(r3)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		out := b.String()
		if out != wants {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants, out)
		}
	})
	t.Run("WithNamespaceAttr", func(t *testing.T) {
		b.Reset()
		wants := `[2022-11-18T21:21:27+01:00] [info] test message [ a_key: value ; b_test_no: 1 ; namespace: [ c_success_rate: 1 ; d_custom: {custom_key 2} ] ]`

		err := h.Handle(r4)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		out := b.String()
		if out != wants {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants, out)
		}
	})
	t.Run("WithReplFn", func(t *testing.T) {
		b.Reset()
		wants := `[2022-11-18T21:21:27+01:00] [info] test message [ a_key: val ]`

		newH := h.WithReplaceFn(func(a attr.Attr) attr.Attr {
			if a.Key() == "a_key" {
				return a.WithValue("val")
			}
			return a
		})
		err := newH.Handle(r2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		out := b.String()
		if out != wants {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants, out)
		}
	})
}

func TestEnabled(t *testing.T) {
	b := &bytes.Buffer{}
	h := New(b)
	t.Run("LevelSkip", func(t *testing.T) {
		newH := h.WithLevel(level.Warn)

		if newH.Enabled(level.Trace) {
			t.Errorf("expected Trace level to be disabled")
		}
	})
	t.Run("LevelAllowed", func(t *testing.T) {
		newH := h.WithLevel(level.Trace)

		if !newH.Enabled(level.Trace) {
			t.Errorf("expected Trace level to be enabled")
		}
	})
	t.Run("LevelRefUnset", func(t *testing.T) {
		if !h.Enabled(level.Trace) {
			t.Errorf("expected Trace level to be enabled")
		}
	})
}

func TestWithSource(t *testing.T) {
	b := &bytes.Buffer{}
	h := New(b)
	t.Run("True", func(t *testing.T) {
		new := h.WithSource(true)

		if !new.(textHandler).addSource {
			t.Errorf("expected addSource to be true")
		}
	})
}
