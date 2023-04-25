package data

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/osc"
)

func TestJunk(t *testing.T) {
	input := []byte("some junk data")

	junkHeader := &ChunkHeader{}
	junk := &ChunkJunk{
		ChunkHeader: junkHeader,
		Depth:       16,
	}

	t.Run("ParseAndBytes", func(t *testing.T) {
		// clear Subchunk2Size
		junk.Subchunk2Size = 0
		junk.Parse(input)
		junk.ParseFloat(nil)

		output := junk.Bytes()
		if !bytes.Equal(input, output) {
			t.Errorf("output mismatch error: wanted %v ; got %v", input, output)
		}
	})

	t.Run("Value", func(t *testing.T) {
		value := junk.Value()
		for i := range input {
			if int(input[i]) != value[i] {
				t.Errorf("output mismatch error on index #%v: wanted %v ; got %v", i, input[i], value[i])
			}
		}
	})

	t.Run("Float", func(t *testing.T) {
		f := junk.Float()
		if f != nil {
			t.Errorf("output mismatch error: wanted %v ; got %v", nil, f)
		}
	})

	t.Run("ParseSecondRun", func(t *testing.T) {
		// second run to test Parse on a dirty state
		junk.Parse(input)
	})

	t.Run("ChunkHeader", func(t *testing.T) {
		header := junk.Header()
		if !reflect.DeepEqual(junkHeader, header) {
			t.Errorf("output mismatch error: wanted %v ; got %v", junkHeader, header)
		}
	})

	t.Run("BitDepth", func(t *testing.T) {
		depth := junk.BitDepth()
		if depth != 16 {
			t.Errorf("output mismatch error: wanted %v ; got %v", 16, depth)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		junk.Reset()

		if len(junk.Data) != 0 {
			t.Errorf("output mismatch error: wanted %v ; got %v", 0, len(junk.Data))
		}
	})

	t.Run("Generate", func(t *testing.T) {
		junk.Generate(osc.SineWave, 2000, 16, 500*time.Millisecond)

		if len(junk.Data) != 0 {
			t.Errorf("output mismatch error: wanted %v ; got %v", 0, len(junk.Data))
		}
	})
}
