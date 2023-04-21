package data

import "testing"

func TestCopy24to32(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		input := []byte{12, 34, 56, 78, 90, 123}
		wants := []byte{12, 34, 56, 0, 78, 90, 123, 0}

		out := copy24to32(input)

		if len(out) != len(wants) {
			t.Errorf("output length mismatch: wanted %v ; got %v", len(wants), len(out))
		}

		for i := range wants {
			if out[i] != wants[i] {
				t.Errorf("output mismatch on idx #%d: wanted %v ; got %v", i, wants[i], out[i])
				return
			}
		}
	})
}

func BenchmarkCopy24to32(b *testing.B) {
	var out []byte

	for i := 0; i < b.N; i++ {
		out = copy24to32([]byte{12, 34, 56, 78, 90, 123})
	}
	_ = out
}

func TestAppend2Bytes(t *testing.T) {
	t.Run("IndexZero", func(t *testing.T) {
		var (
			wants = []byte{12, 34}
			input = [2]byte{12, 34}
			idx   = 0
			out   = []byte{0, 0}
		)

		append2Bytes(idx, out, input)

		if len(out) != len(wants) {
			t.Errorf("output length mismatch: wanted %v ; got %v", len(wants), len(out))
		}

		for i := range wants {
			if out[i] != wants[i] {
				t.Errorf("output mismatch on idx #%d: wanted %v ; got %v", i, wants[i], out[i])
				return
			}
		}
	})
	t.Run("IndexOne", func(t *testing.T) {
		var (
			wants = []byte{90, 90, 12, 34}
			input = [2]byte{12, 34}
			idx   = 1
			out   = []byte{90, 90, 0, 0}
		)

		append2Bytes(idx, out, input)

		if len(out) != len(wants) {
			t.Errorf("output length mismatch: wanted %v ; got %v", len(wants), len(out))
		}

		for i := range wants {
			if out[i] != wants[i] {
				t.Errorf("output mismatch on idx #%d: wanted %v ; got %v", i, wants[i], out[i])
				return
			}
		}
	})
	t.Run("Overflow", func(t *testing.T) {
		var (
			wants = []byte{90, 90, 0, 0}
			input = [2]byte{12, 34}
			idx   = 10
			out   = []byte{90, 90, 0, 0}
		)

		append2Bytes(idx, out, input)

		if len(out) != len(wants) {
			t.Errorf("output length mismatch: wanted %v ; got %v", len(wants), len(out))
		}

		for i := range wants {
			if out[i] != wants[i] {
				t.Errorf("output mismatch on idx #%d: wanted %v ; got %v", i, wants[i], out[i])
				return
			}
		}
	})
}

func TestAppend3Bytes(t *testing.T) {
	t.Run("IndexZero", func(t *testing.T) {
		var (
			wants = []byte{12, 34, 56}
			input = [3]byte{12, 34, 56}
			idx   = 0
			out   = []byte{0, 0, 0}
		)

		append3Bytes(idx, out, input)

		if len(out) != len(wants) {
			t.Errorf("output length mismatch: wanted %v ; got %v", len(wants), len(out))
		}

		for i := range wants {
			if out[i] != wants[i] {
				t.Errorf("output mismatch on idx #%d: wanted %v ; got %v", i, wants[i], out[i])
				return
			}
		}
	})
	t.Run("IndexOne", func(t *testing.T) {
		var (
			wants = []byte{90, 90, 90, 12, 34, 56}
			input = [3]byte{12, 34, 56}
			idx   = 1
			out   = []byte{90, 90, 90, 0, 0, 0}
		)

		append3Bytes(idx, out, input)

		if len(out) != len(wants) {
			t.Errorf("output length mismatch: wanted %v ; got %v", len(wants), len(out))
		}

		for i := range wants {
			if out[i] != wants[i] {
				t.Errorf("output mismatch on idx #%d: wanted %v ; got %v", i, wants[i], out[i])
				return
			}
		}
	})
	t.Run("Overflow", func(t *testing.T) {
		var (
			wants = []byte{90, 90, 90, 0, 0, 0}
			input = [3]byte{12, 34, 56}
			idx   = 10
			out   = []byte{90, 90, 90, 0, 0, 0}
		)

		append3Bytes(idx, out, input)

		if len(out) != len(wants) {
			t.Errorf("output length mismatch: wanted %v ; got %v", len(wants), len(out))
		}

		for i := range wants {
			if out[i] != wants[i] {
				t.Errorf("output mismatch on idx #%d: wanted %v ; got %v", i, wants[i], out[i])
				return
			}
		}
	})
}

func TestAppend4Bytes(t *testing.T) {
	t.Run("IndexZero", func(t *testing.T) {
		var (
			wants = []byte{12, 34, 56, 78}
			input = [4]byte{12, 34, 56, 78}
			idx   = 0
			out   = []byte{0, 0, 0, 0}
		)

		append4Bytes(idx, out, input)

		if len(out) != len(wants) {
			t.Errorf("output length mismatch: wanted %v ; got %v", len(wants), len(out))
		}

		for i := range wants {
			if out[i] != wants[i] {
				t.Errorf("output mismatch on idx #%d: wanted %v ; got %v", i, wants[i], out[i])
				return
			}
		}
	})
	t.Run("IndexOne", func(t *testing.T) {
		var (
			wants = []byte{90, 90, 90, 90, 12, 34, 56, 78}
			input = [4]byte{12, 34, 56, 78}
			idx   = 1
			out   = []byte{90, 90, 90, 90, 0, 0, 0, 0}
		)

		append4Bytes(idx, out, input)

		if len(out) != len(wants) {
			t.Errorf("output length mismatch: wanted %v ; got %v", len(wants), len(out))
		}

		for i := range wants {
			if out[i] != wants[i] {
				t.Errorf("output mismatch on idx #%d: wanted %v ; got %v", i, wants[i], out[i])
				return
			}
		}
	})
	t.Run("Overflow", func(t *testing.T) {
		var (
			wants = []byte{90, 90, 90, 90, 0, 0, 0, 0}
			input = [4]byte{12, 34, 56, 78}
			idx   = 10
			out   = []byte{90, 90, 90, 90, 0, 0, 0, 0}
		)

		append4Bytes(idx, out, input)

		if len(out) != len(wants) {
			t.Errorf("output length mismatch: wanted %v ; got %v", len(wants), len(out))
		}

		for i := range wants {
			if out[i] != wants[i] {
				t.Errorf("output mismatch on idx #%d: wanted %v ; got %v", i, wants[i], out[i])
				return
			}
		}
	})
}
