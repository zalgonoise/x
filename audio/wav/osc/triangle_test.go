package osc_test

import (
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/wav/osc"

	"github.com/zalgonoise/x/audio/wav"
)

func TestTriangle(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			chunk := wav.NewChunk(16, nil)
			chunk.Generate(osc.TriangleWave, 2000, 44100, time.Millisecond*500)
			if len(chunk.Value()) == 0 {
				t.Errorf("expected chunk data to be generated")
			}
			t.Logf("%+v", chunk.Value()[:1024])
		},
	)
}
