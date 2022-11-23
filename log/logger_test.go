package log

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/handlers/jsonh"
	"github.com/zalgonoise/x/log/handlers/texth"
	"github.com/zalgonoise/x/log/level"
)

func TestNewLogger(t *testing.T) {
	b := &bytes.Buffer{}
	h := jsonh.New(b)
	t.Run("Success", func(t *testing.T) {
		out := New(h)

		if out == nil {
			t.Error("expected output not to be nil")
		}
	})
	t.Run("Fail", func(t *testing.T) {
		out := New(nil)
		if out != nil {
			t.Errorf("expected output to be nil; got %v", out)
		}
	})

}

func TestDefault(t *testing.T) {
	wants := std
	out := Default()

	if !reflect.DeepEqual(wants, out) {
		t.Errorf("unexpected output error: wanted %v ; got %v", wants, out)
	}
}

func TestWith(t *testing.T) {
	l := std

	a1 := []attr.Attr{
		attr.New("a", 1),
	}

	t.Run("WithAttr", func(t *testing.T) {
		out := With(a1...)

		if len(out.(*logger).attrs) != 1 {
			t.Errorf("unexpected attributes length: %v", len(out.(*logger).attrs))
			return
		}
		if !reflect.DeepEqual(a1[0], out.(*logger).attrs[0]) {
			t.Errorf("output mismatch error: wanted %v ; got %v", a1[0], out.(*logger).attrs[0])
		}
		if !reflect.DeepEqual(l.(*logger).h, out.(*logger).h) {
			t.Errorf("output mismatch error: wanted %v ; got %v", l.(*logger).h, out.(*logger).h)
		}
	})
	t.Run("WithoutAttr", func(t *testing.T) {
		out := With()

		if !reflect.DeepEqual(l, out) {
			t.Errorf("output mismatch error: wanted %v ; got %v", l, out)
		}
	})
}

func TestLoggerWith(t *testing.T) {
	b := &bytes.Buffer{}
	h := jsonh.New(b)

	a1 := []attr.Attr{
		attr.New("a", 1),
	}
	a2 := []attr.Attr{
		attr.New("key", "value"),
		attr.New("test", true),
	}

	t.Run("Replace", func(t *testing.T) {
		l := New(h)
		out := l.With(a1...)

		if len(out.(*logger).attrs) != 1 {
			t.Errorf("unexpected attributes length: %v", len(out.(*logger).attrs))
			return
		}
		if !reflect.DeepEqual(a1[0], out.(*logger).attrs[0]) {
			t.Errorf("output mismatch error: wanted %v ; got %v", a1[0], out.(*logger).attrs[0])
		}
	})
	t.Run("Erase", func(t *testing.T) {
		l := New(h).With(a2...)

		if len(l.(*logger).attrs) != 2 {
			t.Errorf("unexpected attributes length: %v", len(l.(*logger).attrs))
			return
		}

		out := l.With()
		if len(out.(*logger).attrs) != 0 {
			t.Errorf("unexpected attributes length: %v", len(l.(*logger).attrs))
			return
		}
	})
}

func TestLoggerEnabled(t *testing.T) {
	b := &bytes.Buffer{}
	h := jsonh.New(b)
	hWarn := h.WithLevel(level.Warn)

	t.Run("Default", func(t *testing.T) {
		out := New(h)

		if !out.Enabled(level.Trace) {
			t.Errorf("expected default logger to accept all levels")
		}
	})
	t.Run("FilterWarn", func(t *testing.T) {
		out := New(hWarn)

		if out.Enabled(level.Info) {
			t.Errorf("expected default logger to accept only accept Warn and above")
		}
	})
	t.Run("NilInput", func(t *testing.T) {
		out := New(hWarn)

		if !out.Enabled(nil) {
			t.Errorf("expected nil input to return true")
		}
	})
}

func TestLoggerHandler(t *testing.T) {
	b := &bytes.Buffer{}
	jh := jsonh.New(b)
	th := texth.New(b)

	t.Run("JSONHandler", func(t *testing.T) {
		l := New(jh)

		out := l.Handler()

		if !reflect.DeepEqual(jh, out) {
			t.Errorf("output mismatch error: wanted %v ; got %v", jh, out)
		}
	})
	t.Run("TextHandler", func(t *testing.T) {
		l := New(th)

		out := l.Handler()

		if !reflect.DeepEqual(th, out) {
			t.Errorf("output mismatch error: wanted %v ; got %v", th, out)
		}
	})
}

func TestLoggerLog(t *testing.T) {

}

func TestLoggerTrace(t *testing.T) {

}

func TestLoggerDebug(t *testing.T) {

}

func TestLoggerInfo(t *testing.T) {

}

func TestLoggerWarn(t *testing.T) {

}

func TestLoggerError(t *testing.T) {

}

func TestLoggerFatal(t *testing.T) {

}
