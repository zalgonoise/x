package prometheus_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio"
	promexp "github.com/zalgonoise/x/audio/contrib/prometheus"
	"github.com/zalgonoise/x/audio/dsp/feature"
	"github.com/zalgonoise/x/audio/dsp/spectral"
	audiotest "github.com/zalgonoise/x/audio/internal/testutil"
	"github.com/zalgonoise/x/audio/sdk/metric"
)

// Review of the spectrum code (2026-11-02).

// A new spectrum layout (a restart at another sample rate) replaces the old one: bands the new stream cannot hold stop
// being reported with the old stream's values.
func TestSpectrumSeriesFollowTheLayout(t *testing.T) {
	reader := metric.NewManualReader()

	bands := []spectral.Band{
		{Low: 1000, Centre: 1500, High: 2000},
		{Low: 2000, Centre: 3000, High: 4000},
		{Low: 5000, Centre: 6500, High: 8000},
	}

	spectrum, err := metric.NewSpectrum("audio.spectrum", metric.PeakHold, metric.WithBands(bands),
		metric.WithFrameSize(8), metric.WithHop(8), metric.WithWindow(rectangular))
	require.NoError(t, err)

	m, err := metric.NewMeter(reader, spectrum)
	require.NoError(t, err)

	reg := prometheus.NewPedanticRegistry()
	require.NoError(t, reg.Register(promexp.NewCollector(reader)))

	// 16 kHz, 8 points: bins every 2 kHz; a 6 kHz cosine (bin 3) reads 0.5 on the 5-8 kHz band
	fast := audio.Format{SampleRate: 16000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: fast, Samples: audiotest.BinCosine(8, 3, 0.5)}))

	got := gather(t, reg)
	require.InDelta(t, 0.5, got[`audio_spectrum{channel="0",frequency="6500"}`], 1e-12)

	// the stream restarts at 8 kHz: bins every 1 kHz up to 4 kHz, so the 5-8 kHz band holds no bin
	slow := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: slow, Samples: make([]float64, 8)}))

	got = gather(t, reg)
	require.Contains(t, got, `audio_spectrum{channel="0",frequency="1500"}`)
	require.NotContains(t, got, `audio_spectrum{channel="0",frequency="6500"}`)

	// both layouts in one delta: the later one wins
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: fast, Samples: audiotest.BinCosine(8, 3, 0.5)}))
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: slow, Samples: make([]float64, 8)}))
	require.NotContains(t, gather(t, reg), `audio_spectrum{channel="0",frequency="6500"}`)
}

// Channels a stream no longer has stop being reported once the instrument reports again.
func TestSeriesFollowTheChannels(t *testing.T) {
	reader := metric.NewManualReader()

	peak, err := metric.NewInstrument("audio.peak", func(x []float64, _ int) float64 { return feature.Peak(x) }, metric.Max)
	require.NoError(t, err)

	m, err := metric.NewMeter(reader, peak)
	require.NoError(t, err)

	reg := prometheus.NewPedanticRegistry()
	require.NoError(t, reg.Register(promexp.NewCollector(reader)))

	stereo := audio.Format{SampleRate: 8000, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM}
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: stereo, Samples: []float64{0.5, 0.25}}))
	require.Len(t, gather(t, reg), 2)
	require.Len(t, gather(t, reg), 2, "a scrape without audio keeps the last values")

	mono := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: mono, Samples: []float64{0.125}}))
	require.Equal(t, map[string]float64{`audio_peak{channel="0"}`: 0.125}, gather(t, reg))
}

// Two instruments whose names sanitise alike fail the scrape instead of overwriting each other.
func TestScalarNameConflict(t *testing.T) {
	reader := metric.NewManualReader()

	peak, err := metric.NewInstrument("audio.peak", func(x []float64, _ int) float64 { return feature.Peak(x) }, metric.Max)
	require.NoError(t, err)

	dc, err := metric.NewInstrument("audio_peak", func(x []float64, _ int) float64 { return feature.DCOffset(x) },
		metric.Mean)
	require.NoError(t, err)

	m, err := metric.NewMeter(reader, peak, dc)
	require.NoError(t, err)

	f := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: f, Samples: []float64{0.5, -0.5}}))

	reg := prometheus.NewPedanticRegistry()
	require.NoError(t, reg.Register(promexp.NewCollector(reader)))

	_, err = reg.Gather()
	require.ErrorIs(t, err, promexp.ErrNameConflict)
}
