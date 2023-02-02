package conv

import "testing"

func BenchmarkToFrom64(b *testing.B) {
	var f float64 = 1.1254964682342
	var to []byte = Float64To(f)

	b.Run("To", func(b *testing.B) {
		var r []byte
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r = Float64To(f)
		}
		_ = r
	})

	b.Run("From", func(b *testing.B) {
		var r float64
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r = Float64From(to)
		}
		_ = r
	})

	b.Run("ToFrom", func(b *testing.B) {
		var r float64
		var buf []byte
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = Float64To(f)
			r = Float64From(buf)
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
		r := Float64From(Float64To(v))

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
		r := Float32From(Float32To(v))

		if r != v {
			t.Errorf("output mismatch error: wanted %f ; got %f", v, r)
		}
	}

}
