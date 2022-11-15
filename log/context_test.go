package log

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/log/handlers/jsonh"
)

func TestInContext(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		b := &bytes.Buffer{}
		wants := New(jsonh.New(b))
		input := InContext(context.Background(), wants)

		v := input.Value(StandardCtxKey)
		if v == nil {
			t.Error("output is unexpectedly nil")
		}
		var (
			out Logger
			ok  bool
		)
		if out, ok = v.(Logger); !ok {
			t.Errorf("output is not a Logger interface")
		}

		if !reflect.DeepEqual(wants, out) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, out)
		}
	})
	t.Run("FailNoContext", func(t *testing.T) {
		b := &bytes.Buffer{}
		l := New(jsonh.New(b))
		out := InContext(nil, l)

		if out != nil {
			t.Errorf("output is not nil")
		}
	})
	t.Run("FailNoLogger", func(t *testing.T) {
		out := InContext(context.Background(), nil)

		if out != nil {
			t.Errorf("output is not nil")
		}
	})
	t.Run("FailNoInput", func(t *testing.T) {
		out := InContext(nil, nil)

		if out != nil {
			t.Errorf("output is not nil")
		}
	})
}

func TestFrom(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		b := &bytes.Buffer{}
		wants := New(jsonh.New(b))
		input := context.WithValue(context.Background(), StandardCtxKey, wants)

		out := From(input)
		if !reflect.DeepEqual(wants, out) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, out)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		input := context.Background()

		out := From(input)
		if out != nil {
			t.Errorf("output mismatch error: wanted %v ; got %v", nil, out)
		}
	})
}
