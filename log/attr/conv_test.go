package attr

import "testing"

func TestInt(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		var input int = 1
		var wants int64 = 1

		a := Int("int", input)

		v, ok := a.Value().(int64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("int8", func(t *testing.T) {
		var input int8 = 1
		var wants int64 = 1

		a := Int("int", input)

		v, ok := a.Value().(int64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("int16", func(t *testing.T) {
		var input int16 = 1
		var wants int64 = 1

		a := Int("int", input)

		v, ok := a.Value().(int64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("int32", func(t *testing.T) {
		var input int32 = 1
		var wants int64 = 1

		a := Int("int", input)

		v, ok := a.Value().(int64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("int64", func(t *testing.T) {
		var input int64 = 1
		var wants int64 = 1

		a := Int("int", input)

		v, ok := a.Value().(int64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("custom", func(t *testing.T) {
		type custom int8
		var input custom = 1
		var wants int64 = 1

		a := Int("int", input)

		v, ok := a.Value().(int64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
}

func TestUint(t *testing.T) {
	t.Run("uint", func(t *testing.T) {
		var input uint = 1
		var wants uint64 = 1

		a := Uint("uint", input)

		v, ok := a.Value().(uint64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("uint8", func(t *testing.T) {
		var input uint8 = 1
		var wants uint64 = 1

		a := Uint("uint", input)

		v, ok := a.Value().(uint64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("uint16", func(t *testing.T) {
		var input uint16 = 1
		var wants uint64 = 1

		a := Uint("uint", input)

		v, ok := a.Value().(uint64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("uint32", func(t *testing.T) {
		var input uint32 = 1
		var wants uint64 = 1

		a := Uint("uint", input)

		v, ok := a.Value().(uint64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("uint64", func(t *testing.T) {
		var input uint64 = 1
		var wants uint64 = 1

		a := Uint("uint", input)

		v, ok := a.Value().(uint64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("custom", func(t *testing.T) {
		type custom uint8
		var input custom = 1
		var wants uint64 = 1

		a := Uint("uint", input)

		v, ok := a.Value().(uint64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
}
func TestFloat(t *testing.T) {
	t.Run("float32", func(t *testing.T) {
		var input float32 = 1.0
		var wants float64 = 1.0

		a := Float("float", input)

		v, ok := a.Value().(float64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("float64", func(t *testing.T) {
		var input float64 = 1.0
		var wants float64 = 1.0

		a := Float("float", input)

		v, ok := a.Value().(float64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("custom", func(t *testing.T) {
		type custom float32
		var input custom = 1.0
		var wants float64 = 1.0

		a := Float("float", input)

		v, ok := a.Value().(float64)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
}

func TestComplex(t *testing.T) {
	t.Run("complex64", func(t *testing.T) {
		var input complex64 = 1 + 0.6i
		var wants complex128 = (complex128)((complex64)(1 + 0.6i))

		a := Complex("complex", input)

		v, ok := a.Value().(complex128)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("complex128", func(t *testing.T) {
		var input complex128 = 1 + 0.6i
		var wants complex128 = 1 + 0.6i

		a := Complex("complex", input)

		v, ok := a.Value().(complex128)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("custom", func(t *testing.T) {
		type custom complex64
		var input custom = 1 + 0.6i
		var wants complex128 = (complex128)((custom)(1 + 0.6i))

		a := Complex("complex", input)

		v, ok := a.Value().(complex128)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
}

func TestString(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var input string = "text"
		var wants string = "text"

		a := String("text", input)

		v, ok := a.Value().(string)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("byte", func(t *testing.T) {
		var input byte = 64
		var wants string = "@"

		a := String("text", input)

		v, ok := a.Value().(string)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("rune", func(t *testing.T) {
		var input rune = '@'
		var wants string = "@"

		a := String("text", input)

		v, ok := a.Value().(string)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("[]byte", func(t *testing.T) {
		var input []byte = []byte("text")
		var wants string = "text"

		a := String("text", input)

		v, ok := a.Value().(string)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("[]rune", func(t *testing.T) {
		var input []rune = []rune{'t', 'e', 'x', 't'}
		var wants string = "text"

		a := String("text", input)

		v, ok := a.Value().(string)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
	t.Run("custom", func(t *testing.T) {
		type custom string
		var input custom = "text"
		var wants string = "text"

		a := String("text", input)

		v, ok := a.Value().(string)
		if !ok {
			t.Errorf("expected %T type, got %T", wants, a.Value())
		}
		if v != wants {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, a.Value())
		}
	})
}
