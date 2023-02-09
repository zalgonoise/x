package enc

import (
	"strings"
	"testing"
)

//	message Person {
//	    string name = 2;
//	    uint64 age = 3;
//	    uint64 is_admin = 4;
//	    uint64 id = 5;
//	}
func printHeaders(t *testing.T) {
	ids := []IDAndWire{
		{2, 2, "name"},
		{3, 0, "age"},
		{4, 0, "id"},
		{5, 0, "isAdmin"},
	}
	encodeVarint := func(value uint64) []byte {
		i := 0
		out := make([]byte, 0, 10)
		for value >= 0x80 {
			out = append(out, byte(value)|0x80)
			value >>= 7
			i++
		}
		out = append(out, byte(value))
		return out
	}

	args := []any{}
	s := new(strings.Builder)
	s.WriteString("\n")
	for i := 0; i < 4; i++ {
		s.WriteString("ID: %d, Type: %d, Val: %v, Bin: %08b\t")
		args = append(args, ids[i].ID)
		args = append(args, ids[i].Wire)
		byt := encodeVarint(uint64((ids[i].ID << 3) | ids[i].Wire))
		args = append(args, byt)
		args = append(args, byt)
	}
	t.Logf(s.String(), args...)
}

func printBin(t *testing.T, data []byte) {
	s := new(strings.Builder)
	s.WriteString("\n")
	args := []any{}

	var i int
	for _, b := range data {
		if i > 3 {
			s.WriteString("\n")
			i = 0
		}
		s.WriteString("%08b\t")
		i++
		args = append(args, b)
	}
	t.Logf(s.String(), args...)
}

func TestEncode(t *testing.T) {
	headers := []IDAndWire{
		{2, 2, "name"},
		{3, 0, "age"},
		{4, 0, "id"},
		{5, 0, "isAdmin"},
	}

	b := NewEncoder(0)

	b.EncodeField(2, 2, []byte("pb by hand"))
	b.EncodeVarintField(3, 30)
	b.EncodeVarintField(4, 1)
	b.EncodeVarintField(5, 103)

	t.Log(b.String(), b.Bytes())
	buf := b.Bytes()
	t.Log(HeaderGoString(headers...))
	printBin(t, buf)

	d := NewDecoder(buf)
	out, err := d.Decode()
	if err != nil {
		t.Error(err)
	}
	if out.Name != "pb by hand" ||
		out.Age != 30 ||
		out.IsAdmin != 1 ||
		out.ID != 103 {
		t.Error("invalid output:", out)
	}
	t.Log(out.Name, out.Age, out.ID, out.IsAdmin)
	// t.Error()
}

func BenchmarkEncodeDecode(b *testing.B) {
	b.Run("TypeEncode", func(b *testing.B) {
		p := Person{
			Name:    "pb by hand",
			Age:     30,
			IsAdmin: 1,
			ID:      103,
		}

		var buf []byte

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = p.Bytes()
		}
		_ = buf
	})
	b.Run("TypeEncodeLarge", func(b *testing.B) {
		p := Person{
			Name:    "pb by hand pb by hand pb by hand pb by hand pb by hand pb by hand pb by hand pb by hand pb by hand pb by hand",
			Age:     12312565464576345234,
			IsAdmin: 12546457314654633453,
			ID:      16754345645324267681,
		}

		var buf []byte

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = p.Bytes()
		}
		_ = buf
	})
	b.Run("TypeDecode", func(b *testing.B) {
		p := Person{
			Name:    "pb by hand",
			Age:     30,
			IsAdmin: 1,
			ID:      103,
		}

		buf := p.Bytes()
		var newP Person
		var err error

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dec := NewDecoder(buf)
			newP, err = dec.Decode()
			if err != nil {
				b.Error(err)
				return
			}
		}
		_ = newP
	})
	b.Run("TypeEncodeDecode", func(b *testing.B) {
		p := Person{
			Name:    "pb by hand",
			Age:     30,
			IsAdmin: 1,
			ID:      103,
		}

		var newP Person
		var err error

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := p.Bytes()
			dec := NewDecoder(buf)
			newP, err = dec.Decode()
			if err != nil {
				b.Error(err)
				return
			}
		}
		_ = newP
	})
}

func BenchmarkEncoding(b *testing.B) {

	b.Run("Encode", func(b *testing.B) {
		b.Run("SimpleVarint", func(b *testing.B) {
			enc := NewEncoder(0)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				enc.EncodeVarintField(5, 103)
			}
			b.StopTimer()
			_ = enc.String()
		})
		b.Run("SimpleLen10Bytes", func(b *testing.B) {
			enc := NewEncoder(0)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				enc.EncodeField(2, 2, []byte("pb by hand"))
			}
			b.StopTimer()
			_ = enc.String()
		})
		b.Run("MultiField", func(b *testing.B) {
			enc := NewEncoder(0)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				enc.EncodeField(2, 2, []byte("pb by hand"))
				enc.EncodeVarintField(3, 30)
				enc.EncodeVarintField(4, 1)
				enc.EncodeVarintField(5, 103)
			}
			b.StopTimer()
			_ = enc.String()
		})

	})
	b.Run("Decode", func(b *testing.B) {
		b.Run("SimpleVarint", func(b *testing.B) {
			var wants uint64 = 103
			var input = []byte{40, 103}
			var err error
			var p Person

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dec := NewDecoder(input)
				p, err = dec.Decode()
				if err != nil {
					b.Error(err)
					return
				}
			}
			b.StopTimer()

			// verify
			if p.ID != wants {
				b.Errorf("unexpected output error: wanted %d ; got %d", p.ID, wants)
			}
		})
		b.Run("SimpleLen10Bytes", func(b *testing.B) {
			var wants string = "pb by hand"
			var input = []byte{18, 10, 112, 98, 32, 98, 121, 32, 104, 97, 110, 100}
			var p Person
			var err error

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dec := NewDecoder(input)
				p, err = dec.Decode()
				if err != nil {
					b.Error(err)
					return
				}
			}
			b.StopTimer()

			// verify
			if p.Name != wants {
				b.Errorf("unexpected output error: wanted %s ; got %s", p.Name, wants)
			}
		})
		b.Run("MultiField", func(b *testing.B) {
			var wants = Person{
				Name:    "pb by hand",
				Age:     30,
				IsAdmin: 1,
				ID:      103,
			}
			var input = []byte{18, 10, 112, 98, 32, 98, 121, 32, 104, 97, 110, 100, 24, 30, 32, 1, 40, 103}
			var p Person
			var err error

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dec := NewDecoder(input)
				p, err = dec.Decode()
				if err != nil {
					b.Error(err)
					return
				}
			}
			b.StopTimer()

			// verify
			if p.Name != wants.Name {
				b.Errorf("unexpected output error: wanted %s ; got %s", wants.Name, p.Name)
			}
			if p.ID != wants.ID {
				b.Errorf("unexpected output error: wanted %d ; got %d", wants.ID, p.ID)
			}
			if p.Age != wants.Age {
				b.Errorf("unexpected output error: wanted %d ; got %d", wants.Age, p.Age)
			}
			if p.IsAdmin != wants.IsAdmin {
				b.Errorf("unexpected output error: wanted %d ; got %d", wants.IsAdmin, p.IsAdmin)
			}
		})
	})
}
