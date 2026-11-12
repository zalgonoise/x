package exporters_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/sdk"
	"github.com/zalgonoise/x/audio/sdk/exporters"
)

func TestWAV(t *testing.T) {
	f := audio.Format{SampleRate: 8000, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM}

	var out bytes.Buffer

	e := exporters.NewWAV(&out)
	require.NoError(t, e.Export(t.Context(), audio.Buffer{Format: f, Samples: []float64{0.5, -0.5}}))
	require.NoError(t, e.Export(t.Context(), audio.Buffer{Format: f, Samples: []float64{0.25, -0.25}, Frame: 1}))
	require.NoError(t, e.ForceFlush(t.Context()))

	other := f
	other.SampleRate = 16000
	require.ErrorIs(t, e.Export(t.Context(), audio.Buffer{Format: other, Samples: []float64{0, 0}}), exporters.ErrFormatChanged)

	require.NoError(t, e.Shutdown(t.Context()))
	require.NoError(t, e.Shutdown(t.Context()), "idempotent")
	require.ErrorIs(t, e.Export(t.Context(), audio.Buffer{Format: f, Samples: []float64{0, 0}}), sdk.ErrShutdown)
	require.ErrorIs(t, e.ForceFlush(t.Context()), sdk.ErrShutdown)

	d, err := wav.NewDecoder(bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	require.Equal(t, f, d.Format())

	got, err := audio.ReadAll(d)
	require.NoError(t, err)
	require.Equal(t, []float64{0.5, -0.5, 0.25, -0.25}, got.Samples)

	var empty bytes.Buffer

	require.NoError(t, exporters.NewWAV(&empty).Shutdown(t.Context()))
	require.Zero(t, empty.Len(), "nothing exported, nothing written")

	bad := exporters.NewWAV(&empty)
	require.ErrorIs(t, bad.Export(t.Context(), audio.Buffer{Format: audio.Format{}, Samples: []float64{0}}), audio.ErrInvalidFormat)
}
