package types

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
	b.EncodeVarintField(3, 301)
	b.EncodeVarintField(4, 1)

	t.Log(b.String(), b.Bytes())
	buf := b.Bytes()
	printBin(t, buf)

	d := NewDecoder(buf)
	out, err := d.Decode()
	if err != nil {
		t.Error(err)
	}

	if len(out) == 0 {
		t.Error("EMPTY MAP")
		return
	}

	t.Log(out, out[5], out[5].Value(), out[5].Num())
	t.Log(out, out[2], string((out[2].Value()).([]byte)), out[2].Num())
	t.Log(out, out[3], out[3].Value(), out[3].Num())
	t.Log(out, out[4], out[4].Value(), out[4].Num())
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
			var f map[int]Field
			var err error

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dec := NewDecoder(input)
				f, err = dec.Decode()
				if err != nil {
					b.Error(err)
					return
				}
			}
			b.StopTimer()

			// verify
			field, ok := f[5]
			if !ok {
				b.Error("no field number 5")
				return
			}
			if (field.Value()).(uint64) != wants {
				b.Errorf("unexpected output error: wanted %d ; got %d", field.Value(), wants)
			}
		})
		b.Run("SimpleLen10Bytes", func(b *testing.B) {
			var wants string = "pb by hand"
			var input = []byte{18, 10, 112, 98, 32, 98, 121, 32, 104, 97, 110, 100}
			var f map[int]Field
			var err error

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dec := NewDecoder(input)
				f, err = dec.Decode()
				if err != nil {
					b.Error(err)
					return
				}
			}
			b.StopTimer()

			// verify
			field, ok := f[2]
			if !ok {
				b.Error("no field number 2")
				return
			}
			if string((field.Value()).([]byte)) != wants {
				b.Errorf("unexpected output error: wanted %s ; got %s", field.Value(), wants)
			}
		})
		b.Run("MultiField", func(b *testing.B) {
			var wantsUint64 = map[int]uint64{
				5: 103,
				3: 301,
				4: 1,
			}
			var wantsBytes = map[int]string{
				2: "pb by hand",
			}
			var input = []byte{40, 103, 18, 10, 112, 98, 32, 98, 121, 32, 104, 97, 110, 100, 24, 173, 2, 32, 1}
			var f map[int]Field
			var err error

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dec := NewDecoder(input)
				f, err = dec.Decode()
				if err != nil {
					b.Error(err)
					return
				}
			}
			b.StopTimer()

			// verify
			for k, v := range wantsBytes {
				field, ok := f[k]
				if !ok {
					b.Errorf("no field number %d", k)
					return
				}
				if string((field.Value()).([]byte)) != v {
					b.Errorf("unexpected output error: wanted %s ; got %s", field.Value(), v)
				}
			}
			for k, v := range wantsUint64 {
				field, ok := f[k]
				if !ok {
					b.Errorf("no field number %d", k)
					return
				}
				if (field.Value()).(uint64) != v {
					b.Errorf("unexpected output error: wanted %d ; got %d", field.Value(), v)
				}
			}
		})
	})
}
