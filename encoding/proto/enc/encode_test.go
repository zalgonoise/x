package enc

import (
	"strings"
	"testing"
)

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
	b := NewEncoder()

	b.EncodeVarintField(5, 103)
	b.EncodeField(2, 2, []byte("pb by hand"))
	b.EncodeVarintField(3, 30)
	b.EncodeVarintField(4, 1)

	t.Log(b.String(), b.Bytes())
	buf := b.Bytes()
	printBin(t, buf)

	d := NewDecoder(buf)
	out, err := d.Decode()
	if err != nil {
		t.Error(err)
	}

	t.Log(out.Name, out.Age, out.ID, out.IsAdmin)
}

func BenchmarkEncoding(b *testing.B) {
	b.Run("Encode", func(b *testing.B) {
		b.Run("SimpleVarint", func(b *testing.B) {
			enc := NewEncoder()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				enc.EncodeVarintField(5, 103)
			}
			b.StopTimer()
			_ = enc.String()
		})
		b.Run("SimpleLen10Bytes", func(b *testing.B) {
			enc := NewEncoder()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				enc.EncodeField(2, 2, []byte("pb by hand"))
			}
			b.StopTimer()
			_ = enc.String()
		})
		b.Run("MultiField", func(b *testing.B) {
			enc := NewEncoder()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				enc.EncodeVarintField(5, 103)
				enc.EncodeField(2, 2, []byte("pb by hand"))
				enc.EncodeVarintField(3, 301)
				enc.EncodeVarintField(4, 1)
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
				ID:      103,
				Age:     30,
				IsAdmin: 1,
				Name:    "pb by hand",
			}
			var input = []byte{40, 103, 18, 10, 112, 98, 32, 98, 121, 32, 104, 97, 110, 100, 24, 30, 32, 1}
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
