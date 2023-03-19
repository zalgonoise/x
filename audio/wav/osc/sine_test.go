package osc_test

import (
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/osc"
)

func TestSine(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			chunk := wav.NewChunk(16, nil)
			chunk.Generate(osc.SineWave, 2000, 44100, time.Second/2)
			if len(chunk.Value()) == 0 {
				t.Errorf("expected chunk data to be generated")
			}
			t.Logf("%+v", chunk.Value()[:1024])
		},
	)
}
