package forms_test

import (
	"testing"

	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/forms"
)

func TestSine(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			chunk := wav.NewChunk(16, nil)
			chunk.Generate(forms.SineWave, 2000, 0.5, 44100)
			if len(chunk.Value()) == 0 {
				t.Errorf("expected chunk data to be generated")
			}
		},
	)
}
