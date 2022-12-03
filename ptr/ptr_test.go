package ptr_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/ptr"
)

func FuzzToAndFrom(f *testing.F) {
	f.Add("test")
	f.Add("pointer")
	f.Add("with strings!")
	f.Add("!@#$%^&*()_+[]{};'\\:|")
	f.Add(`
            

`)

	f.Fuzz(func(t *testing.T, a string) {
		p := ptr.To(a)
		out, ok := ptr.From(p)
		if !ok || a != out {
			t.Errorf("unexpected output: sent %v ; got %v", a, out)
		}
	})
}

func TestFrom(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			wants := "test"
			v := &wants
			p, ok := ptr.From(v)
			if !ok {
				t.Errorf("expected input pointer not to be nil")
			}
			if wants != p {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			var v *string
			p, ok := ptr.From(v)
			if ok {
				t.Errorf("expected input pointer to be nil")
			}
			if p != "" {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
	})
	t.Run("Float", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			wants := 1.5
			v := &wants
			p, ok := ptr.From(v)
			if !ok {
				t.Errorf("expected input pointer not to be nil")
			}
			if wants != p {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			var v *float64
			p, ok := ptr.From(v)
			if ok {
				t.Errorf("expected input pointer to be nil")
			}
			if p != 0.0 {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
	})
	t.Run("Struct", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			var wants = struct {
				name string
				id   int
			}{
				name: "test",
				id:   0,
			}

			v := &wants
			p, ok := ptr.From(v)
			if !ok {
				t.Errorf("expected input pointer not to be nil")
			}
			if !reflect.DeepEqual(wants, p) {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			type card struct {
				name string
				id   int
			}

			var v *card
			p, ok := ptr.From(v)
			if ok {
				t.Errorf("expected input pointer to be nil")
			}
			if !reflect.DeepEqual(card{}, p) {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
	})
	t.Run("StructWithPointers", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			item1 := "test"
			item2 := 0
			var wants = struct {
				name *string
				id   *int
			}{
				name: &item1,
				id:   &item2,
			}

			v := &wants
			p, ok := ptr.From(v)

			if !ok {
				t.Errorf("expected input pointer not to be nil")
			}
			if !reflect.DeepEqual(wants, p) {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			type card struct {
				name *string
				id   *int
			}

			var v *card
			p, ok := ptr.From(v)
			if ok {
				t.Errorf("expected input pointer to be nil")
			}
			if !reflect.DeepEqual(card{}, p) {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
	})
	t.Run("PointerOfPointer", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			wants := "test"
			v1 := &wants
			v2 := &v1
			p, ok := ptr.From(v2)

			if !ok {
				t.Errorf("expected input pointer not to be nil")
			}
			if !reflect.DeepEqual(v1, p) {
				t.Errorf("unexpected output: wanted %v ; got %v", v1, p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			var v **string
			p, ok := ptr.From(v)
			if ok {
				t.Errorf("expected input pointer to be nil")
			}
			if p != nil {
				t.Errorf("expected returned value to be nil")
			}
		})
	})
}

func TestMust(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			wants := "test"
			v := &wants
			p := ptr.Must(v)
			if wants != p {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			var v *string
			p := ptr.Must(v)
			if p != "" {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
	})
	t.Run("Float", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			wants := 1.5
			v := &wants
			p := ptr.Must(v)
			if wants != p {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			var v *float64
			p := ptr.Must(v)
			if p != 0.0 {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
	})
	t.Run("Struct", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			var wants = struct {
				name string
				id   int
			}{
				name: "test",
				id:   0,
			}

			v := &wants
			p := ptr.Must(v)
			if !reflect.DeepEqual(wants, p) {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			type card struct {
				name string
				id   int
			}

			var v *card
			p := ptr.Must(v)
			if !reflect.DeepEqual(card{}, p) {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
	})
	t.Run("StructWithPointers", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			item1 := "test"
			item2 := 0
			var wants = struct {
				name *string
				id   *int
			}{
				name: &item1,
				id:   &item2,
			}

			v := &wants
			p := ptr.Must(v)
			if !reflect.DeepEqual(wants, p) {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			type card struct {
				name *string
				id   *int
			}

			var v *card
			p := ptr.Must(v)
			if !reflect.DeepEqual(card{}, p) {
				t.Errorf("unexpected output: wanted %v ; got %v", v, p)
			}
		})
	})
	t.Run("PointerOfPointer", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			wants := "test"
			v1 := &wants
			v2 := &v1
			p := ptr.Must(v2)
			if !reflect.DeepEqual(v1, p) {
				t.Errorf("unexpected output: wanted %v ; got %v", v1, p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			var v **string
			p := ptr.Must(v)
			if p != nil {
				t.Errorf("expected returned value to be nil")
			}
		})
	})
}

func TestTo(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			wants := "test"
			p := ptr.To(wants)
			if wants != *p {
				t.Errorf("unexpected output: wanted %v ; got %v", wants, *p)
			}
		})
	})
	t.Run("Float", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			wants := 1.5
			p := ptr.To(wants)
			if wants != *p {
				t.Errorf("unexpected output: wanted %v ; got %v", wants, *p)
			}
		})
	})
	t.Run("Struct", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			var wants = struct {
				name string
				id   int
			}{
				name: "test",
				id:   0,
			}

			p := ptr.To(wants)
			if !reflect.DeepEqual(wants, *p) {
				t.Errorf("unexpected output: wanted %v ; got %v", wants, *p)
			}
		})
		t.Run("ZeroValue", func(t *testing.T) {
			wants := struct {
				name string
				id   int
			}{}

			p := ptr.To(wants)
			if !reflect.DeepEqual(wants, *p) {
				t.Errorf("unexpected output: wanted %v ; got %v", wants, *p)
			}
		})
	})
	t.Run("StructWithPointers", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			item1 := "test"
			item2 := 0
			var wants = struct {
				name *string
				id   *int
			}{
				name: &item1,
				id:   &item2,
			}

			p := ptr.To(wants)
			if !reflect.DeepEqual(wants, *p) {
				t.Errorf("unexpected output: wanted %v ; got %v", wants, *p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			var wants = struct {
				name *string
				id   *int
			}{}

			p := ptr.To(wants)
			if !reflect.DeepEqual(wants, *p) {
				t.Errorf("unexpected output: wanted %v ; got %v", wants, *p)
			}
		})
	})
	t.Run("PointerOfPointer", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			wants := "test"
			v := &wants
			p := ptr.To(v)
			if !reflect.DeepEqual(v, *p) {
				t.Errorf("unexpected output: wanted %v ; got %v", v, *p)
			}
		})
		t.Run("NilValue", func(t *testing.T) {
			var v *string
			p := ptr.To(v)
			if p != nil && *p != nil {
				t.Errorf("expected inner value to be nil")
			}
		})
	})
}

func TestCopy(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		var addr1 string
		var addr2 string

		v := "test"
		p := &v
		addr1 = fmt.Sprintf("%p", p)

		c := ptr.Copy(p)
		addr2 = fmt.Sprintf("%p", c)

		if c == p {
			t.Errorf("expected pointer addresses to be different: initial: %s ; copy: %s", addr1, addr2)
		}

		if *c != *p {
			t.Errorf("expected pointer values to be the same: initial: %v ; copy: %v", *p, *c)
		}
	})
	t.Run("Float", func(t *testing.T) {
		var addr1 string
		var addr2 string

		v := 1.5
		p := &v
		addr1 = fmt.Sprintf("%p", p)

		c := ptr.Copy(p)
		addr2 = fmt.Sprintf("%p", c)

		if c == p {
			t.Errorf("expected pointer addresses to be different: initial: %s ; copy: %s", addr1, addr2)
		}

		if *c != *p {
			t.Errorf("expected pointer values to be the same: initial: %v ; copy: %v", *p, *c)
		}
	})
	t.Run("Struct", func(t *testing.T) {
		var addr1 string
		var addr2 string

		v := struct {
			name string
			id   int
		}{
			name: "test",
			id:   0,
		}
		p := &v
		addr1 = fmt.Sprintf("%p", p)

		c := ptr.Copy(p)
		addr2 = fmt.Sprintf("%p", c)

		if c == p {
			t.Errorf("expected pointer addresses to be different: initial: %s ; copy: %s", addr1, addr2)
		}

		if !reflect.DeepEqual(*p, *c) {
			t.Errorf("expected pointer values to be the same: initial: %v ; copy: %v", *p, *c)
		}
	})
	t.Run("StructWithPointers", func(t *testing.T) {
		var addr1 string
		var addr2 string

		data1 := "test"
		data2 := 0
		v := struct {
			name *string
			id   *int
		}{
			name: &data1,
			id:   &data2,
		}
		p := &v
		addr1 = fmt.Sprintf("%p", p)

		c := ptr.Copy(p)
		addr2 = fmt.Sprintf("%p", c)

		if c == p {
			t.Errorf("expected pointer addresses to be different: initial: %s ; copy: %s", addr1, addr2)
		}

		if !reflect.DeepEqual(*p, *c) {
			t.Errorf("expected pointer values to be the same: initial: %v ; copy: %v", *p, *c)
		}
	})
	t.Run("PointerOfPointer", func(t *testing.T) {
		var addr1 string
		var addr2 string

		v1 := "test"
		v2 := &v1
		p := &v2
		addr1 = fmt.Sprintf("%p", p)

		c := ptr.Copy(p)
		addr2 = fmt.Sprintf("%p", c)

		if c == p {
			t.Errorf("expected pointer addresses to be different: initial: %s ; copy: %s", addr1, addr2)
		}

		if !reflect.DeepEqual(**p, **c) {
			t.Errorf("expected pointer values to be the same: initial: %v ; copy: %v", **p, **c)
		}
	})
}
