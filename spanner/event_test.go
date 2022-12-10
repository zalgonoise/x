package spanner

import (
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/attr"
)

func TestNewEvent(t *testing.T) {
	var (
		name  = "test"
		attrs = []attr.Attr{
			attr.String("attr", "attr"),
			attr.Int("idx", 0),
		}
	)

	t.Run("Simple", func(t *testing.T) {
		e := newEvent(name)

		if e.name != name {
			t.Errorf("unexpected output error: wanted %v ; got %v", name, e.name)
		}
		if e.timestamp.IsZero() || reflect.DeepEqual(time.Unix(0, 0), e.timestamp) {
			t.Errorf("invalid time: %v", e.timestamp)
		}
		if len(e.attrs) != 0 {
			t.Errorf("expected empty attribute list")
		}
	})
	t.Run("WithAttrs", func(t *testing.T) {
		e := newEvent(name, attrs...)

		if e.name != name {
			t.Errorf("unexpected output error: wanted %v ; got %v", name, e.name)
		}
		if e.timestamp.IsZero() || reflect.DeepEqual(time.Unix(0, 0), e.timestamp) {
			t.Errorf("invalid time: %v", e.timestamp)
		}
		if len(e.attrs) != 2 {
			t.Errorf("expected empty attribute list")
		}
	})
}
