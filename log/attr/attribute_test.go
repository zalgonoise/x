package attr

import (
	"reflect"
	"testing"
)

type testStringer struct{}

func (testStringer) String() string {
	return "text"
}

func TestNew(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var input string = "text"
		var wants string = "text"

		a := New("text", input)

		v, ok := a.Value().(string)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("stringer", func(t *testing.T) {
		var input = testStringer{}
		var wants string = "text"

		a := New("text", input)

		v, ok := a.Value().(testStringer)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v.String() != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, v.String())
		}
	})
	t.Run("struct", func(t *testing.T) {
		var input = struct {
			id int
		}{
			id: 1,
		}
		var wants int = 1

		a := New("text", input)

		v, ok := a.Value().(struct {
			id int
		})
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v.id != 1 {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, v.id)
		}
	})
}

func TestKey(t *testing.T) {
	tc := []struct {
		attr  Attr
		wants string
		ok    bool
	}{
		{
			attr:  New("key", "value"),
			wants: "key",
			ok:    true,
		}, {
			attr:  New("key2", "value"),
			wants: "key2",
			ok:    true,
		}, {
			attr:  New("k", "value"),
			wants: "k",
			ok:    true,
		}, {
			attr:  New("", "value"),
			wants: "",
		},
	}

	for _, tt := range tc {
		if tt.attr == nil {
			if tt.ok {
				t.Errorf("got nil attribute on an supposedly OK test")
			}
			return
		}
		if tt.attr.Key() != tt.wants {
			t.Errorf("output mismatch error: wanted %s ; got %s", tt.wants, tt.attr.Key())
		}
	}
}

func TestValue(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var input string = "text"
		var wants string = "text"
		a := New("key", input)

		v, ok := a.Value().(string)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("struct", func(t *testing.T) {
		var input = struct {
			id int
		}{
			id: 1,
		}
		a := New("key", input)

		v, ok := a.Value().(struct {
			id int
		})
		if !ok {
			t.Errorf("expected %T type, got %T", input, a.Value())
		}
		if !reflect.DeepEqual(input, v) {
			t.Errorf("unexpected value: wanted %v ; got %v", input, v)
		}
	})
	t.Run("int8", func(t *testing.T) {
		var input int8 = 1
		a := New("key", input)

		v, ok := a.Value().(int8)
		if !ok {
			t.Errorf("expected %T type, got %T", input, a.Value())
		}
		if v != input {
			t.Errorf("unexpected value: wanted %v ; got %v", input, v)
		}
	})
}

func TestWithKey(t *testing.T) {
	tc := []struct {
		attr  Attr
		new   string
		wants string
		ok    bool
	}{
		{
			attr:  New("key", "value"),
			new:   "val",
			wants: "val",
			ok:    true,
		}, {
			attr:  New("key2", "value"),
			new:   "val2",
			wants: "val2",
			ok:    true,
		}, {
			attr:  New("k", "value"),
			new:   "v",
			wants: "v",
			ok:    true,
		}, {
			attr:  New("key", "value"),
			new:   "",
			wants: "",
		},
	}

	for _, tt := range tc {
		a := tt.attr.WithKey(tt.new)
		if a == nil {
			if tt.ok {
				t.Errorf("got nil attribute on an supposedly OK test")
			}
			return
		}
		if a.Key() != tt.wants {
			t.Errorf("output mismatch error: wanted %s ; got %s", tt.wants, a.Key())
		}
	}
}

func TestWithValue(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var input string = "text"
		var wants string = "txet"
		attribute := New("key", input)
		a := attribute.WithValue(wants)

		v, ok := a.Value().(string)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("struct", func(t *testing.T) {
		var input = struct {
			id int
		}{
			id: 1,
		}
		var wants = struct {
			id int
		}{
			id: 2,
		}
		attribute := New("key", input)
		a := attribute.WithValue(wants)

		v, ok := a.Value().(struct {
			id int
		})
		if !ok {
			t.Errorf("expected %T type, got %T", input, a.Value())
		}
		if !reflect.DeepEqual(wants, v) {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, v)
		}
	})
	t.Run("int8", func(t *testing.T) {
		var input int8 = 1
		var wants int8 = 2
		attribute := New("key", input)
		a := attribute.WithValue(wants)

		v, ok := a.Value().(int8)
		if !ok {
			t.Errorf("expected %T type, got %T", input, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, v)
		}
	})
	t.Run("nil", func(t *testing.T) {
		var input int8 = 1
		attribute := New("key", input)
		a := attribute.WithValue(nil)

		if a != nil {
			t.Errorf("expected a to be nil, got %v", a)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		var input int8 = 1
		var new = []float64{1.0, 1.1}
		attribute := New("key", input)
		a := attribute.WithValue(new)

		if a != nil {
			t.Errorf("expected a to be nil, got %v", a)
		}
	})
}
