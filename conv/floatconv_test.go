package conv

import "testing"

func BenchmarkToFrom64(b *testing.B) {
	var f float64 = 1.1254964682342
	var to []byte = To64(f)

	b.Run("To", func(b *testing.B) {
		var r []byte
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r = To64(f)
		}
		_ = r
	})

	b.Run("From", func(b *testing.B) {
		var r float64
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r = From64(to)
		}
		_ = r
	})

	b.Run("ToFrom", func(b *testing.B) {
		var r float64
		var buf []byte
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = To64(f)
			r = From64(buf)
		}
		_ = buf
		_ = r
	})
}

func TestToFrom64(t *testing.T) {

	fl := []float64{
		0,
		1.5,
		1.1231256466453,
		123153453.12312,
		1.1254964682342,
	}

	for _, v := range fl {
		r := From64(To64(v))

		if r != v {
			t.Errorf("output mismatch error: wanted %f ; got %f", v, r)
		}
	}

}

func TestToFrom32(t *testing.T) {

	fl := []float32{
		0,
		1.5,
		1.1231256466453,
		123153453.12312,
		1.1254964682342,
	}

	for _, v := range fl {
		r := From32(To32(v))

		if r != v {
			t.Errorf("output mismatch error: wanted %f ; got %f", v, r)
		}
	}

}
