package log

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/zalgonoise/x/log/handlers/jsonh"
	"github.com/zalgonoise/x/log/level"
)

func TestLoggerLog(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)
		testMsg := "test message"
		testLevel := level.Info

		wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"info"}`)

		l := New(h)
		l.Log(testLevel, testMsg)

		if !wants.MatchString(b.String()) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
		}
	})
	t.Run("NilLevel", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)
		testMsg := "test message"

		wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"info"}`)

		l := New(h)
		l.Log(nil, testMsg)

		if !wants.MatchString(b.String()) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
		}
	})
	t.Run("NoMessage", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)
		testLevel := level.Info

		wants := ""

		l := New(h)
		l.Log(testLevel, "")

		if b.String() != wants {
			t.Errorf("output mismatch error: wanted empty string ; got %s", b.String())
		}
	})

}

func TestLoggerTrace(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)
		testMsg := "test message"

		wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"trace"}`)

		l := New(h)
		l.Trace(testMsg)

		if !wants.MatchString(b.String()) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
		}
	})
	t.Run("NoMessage", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)

		wants := ""

		l := New(h)
		l.Trace("")

		if b.String() != wants {
			t.Errorf("output mismatch error: wanted empty string ; got %s", b.String())
		}
	})
}

func TestLoggerDebug(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {

		b := &bytes.Buffer{}
		h := jsonh.New(b)
		testMsg := "test message"

		wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"debug"}`)

		l := New(h)
		l.Debug(testMsg)

		if !wants.MatchString(b.String()) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
		}
	})
	t.Run("NoMessage", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)

		wants := ""

		l := New(h)
		l.Debug("")

		if b.String() != wants {
			t.Errorf("output mismatch error: wanted empty string ; got %s", b.String())
		}
	})
}

func TestLoggerInfo(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)
		testMsg := "test message"

		wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"info"}`)

		l := New(h)
		l.Info(testMsg)

		if !wants.MatchString(b.String()) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
		}
	})
	t.Run("NoMessage", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)

		wants := ""

		l := New(h)
		l.Info("")

		if b.String() != wants {
			t.Errorf("output mismatch error: wanted empty string ; got %s", b.String())
		}
	})
}

func TestLoggerWarn(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)
		testMsg := "test message"

		wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"warn"}`)

		l := New(h)
		l.Warn(testMsg)

		if !wants.MatchString(b.String()) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
		}
	})
	t.Run("NoMessage", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)

		wants := ""

		l := New(h)
		l.Warn("")

		if b.String() != wants {
			t.Errorf("output mismatch error: wanted empty string ; got %s", b.String())
		}
	})
}

func TestLoggerError(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)
		testMsg := "test message"

		wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"error"}`)

		l := New(h)
		l.Error(testMsg)

		if !wants.MatchString(b.String()) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
		}
	})
	t.Run("NoMessage", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)

		wants := ""

		l := New(h)
		l.Error("")

		if b.String() != wants {
			t.Errorf("output mismatch error: wanted empty string ; got %s", b.String())
		}
	})
}

func TestLoggerFatal(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)
		testMsg := "test message"

		wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"fatal"}`)

		l := New(h)
		l.Fatal(testMsg)

		if !wants.MatchString(b.String()) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
		}
	})
	t.Run("NoMessage", func(t *testing.T) {
		b := &bytes.Buffer{}
		h := jsonh.New(b)

		wants := ""

		l := New(h)
		l.Fatal("")

		if b.String() != wants {
			t.Errorf("output mismatch error: wanted empty string ; got %s", b.String())
		}
	})
}
