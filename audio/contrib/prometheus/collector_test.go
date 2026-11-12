package prometheus_test

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio"
	promexp "github.com/zalgonoise/x/audio/contrib/prometheus"
	"github.com/zalgonoise/x/audio/dsp/feature"
	"github.com/zalgonoise/x/audio/dsp/spectral"
	"github.com/zalgonoise/x/audio/fft/window"
	audiotest "github.com/zalgonoise/x/audio/internal/testutil"
	"github.com/zalgonoise/x/audio/sdk/metric"
)

func TestCollector(t *testing.T) {
	reader := metric.NewManualReader()

	peak, err := metric.NewInstrument("audio.peak", func(x []float64, _ int) float64 { return feature.Peak(x) }, metric.Max)
	require.NoError(t, err)

	dc, err := metric.NewInstrument("audio.dc", func(x []float64, _ int) float64 { return feature.DCOffset(x) },
		metric.Mean, metric.OnMixdown())
	require.NoError(t, err)

	m, err := metric.NewMeter(reader, peak, dc)
	require.NoError(t, err)

	f := audio.Format{SampleRate: 8000, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM}
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: f, Samples: []float64{0.5, -0.25, 0.1, 0.25}}))

	reg := prometheus.NewPedanticRegistry()
	require.NoError(t, reg.Register(promexp.NewCollector(reader, promexp.WithNamespace("radio"))))

	want := `
# HELP radio_audio_peak Audio measurement radio_audio_peak, per channel
# TYPE radio_audio_peak gauge
radio_audio_peak{channel="0"} 0.5
radio_audio_peak{channel="1"} 0.25
# HELP radio_audio_dc Audio measurement radio_audio_dc, per channel
# TYPE radio_audio_dc gauge
radio_audio_dc{channel="mix"} 0.15
`
	require.NoError(t, testutil.GatherAndCompare(reg, strings.NewReader(want)))
	require.NoError(t, testutil.GatherAndCompare(reg, strings.NewReader(want)),
		"a scrape without new audio repeats the last values")

	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: f, Samples: []float64{0.9, 0, 0, 0}}))
	require.NoError(t, testutil.GatherAndCompare(reg, strings.NewReader(`
# HELP radio_audio_peak Audio measurement radio_audio_peak, per channel
# TYPE radio_audio_peak gauge
radio_audio_peak{channel="0"} 0.9
radio_audio_peak{channel="1"} 0
# HELP radio_audio_dc Audio measurement radio_audio_dc, per channel
# TYPE radio_audio_dc gauge
radio_audio_dc{channel="mix"} 0.225
`)), "new audio replaces the values")

	broken := prometheus.NewRegistry()
	require.NoError(t, broken.Register(promexp.NewCollector(metric.NewManualReader()))) // no meter: Collect fails

	_, err = broken.Gather()
	require.ErrorIs(t, err, metric.ErrNoMeter, "collection errors reach the scrape")
}

// Spectra are gauges per channel and frequency band, labelled with the band's nominal centre.
func TestCollectorSpectrum(t *testing.T) {
	reader := metric.NewManualReader()

	bands := []spectral.Band{
		{Low: 500, Centre: 707.5, High: 1000},
		{Low: 1000, Centre: 1500, High: 2000},
		{Low: 2000, Centre: 3000, High: 4000},
	}

	spectrum, err := metric.NewSpectrum("audio.spectrum", metric.PeakHold, metric.WithBands(bands),
		metric.WithFrameSize(8), metric.WithHop(8), metric.WithWindow(rectangular))
	require.NoError(t, err)

	peak, err := metric.NewInstrument("audio.peak", func(x []float64, _ int) float64 { return feature.Peak(x) }, metric.Max)
	require.NoError(t, err)

	m, err := metric.NewMeter(reader, spectrum, peak)
	require.NoError(t, err)

	// 8 points at 8 kHz: bins every 1 kHz; a 1 kHz cosine on the left, a 2 kHz one on the right, both on bins
	f := audio.Format{SampleRate: 8000, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM}
	samples := audiotest.Interleave(audiotest.BinCosine(8, 1, 0.5), audiotest.BinCosine(8, 2, 0.25))
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: f, Samples: samples}))

	reg := prometheus.NewPedanticRegistry()
	require.NoError(t, reg.Register(promexp.NewCollector(reader)))

	// the 500-1000 Hz band holds no bin (bins lie at 0, 1000, 2000... Hz) and is left out
	want := map[string]float64{
		`audio_peak{channel="0"}`:                      0.5,
		`audio_peak{channel="1"}`:                      0.25,
		`audio_spectrum{channel="0",frequency="1500"}`: 0.5,
		`audio_spectrum{channel="0",frequency="3000"}`: 0,
		`audio_spectrum{channel="1",frequency="1500"}`: 0,
		`audio_spectrum{channel="1",frequency="3000"}`: 0.25,
	}

	for range 2 { // a scrape without new audio repeats the last spectrum
		got := gather(t, reg)
		require.Len(t, got, len(want))

		for k, v := range want {
			require.Contains(t, got, k)
			require.InDelta(t, v, got[k], 1e-12, k)
		}
	}

	families, err := reg.Gather()
	require.NoError(t, err)
	require.Equal(t, "Audio spectrum audio_spectrum (linear amplitude, 1 is a full-scale sine), per channel and "+
		"frequency in Hz", families[1].GetHelp())
}

// gather returns every gauge of reg by name and labels, as in the exposition format.
func gather(t *testing.T, reg prometheus.Gatherer) map[string]float64 {
	t.Helper()

	families, err := reg.Gather()
	require.NoError(t, err)

	out := map[string]float64{}

	for _, f := range families {
		for _, m := range f.GetMetric() {
			labels := make([]string, 0, len(m.GetLabel()))
			for _, l := range m.GetLabel() {
				labels = append(labels, l.GetName()+`="`+l.GetValue()+`"`)
			}

			out[f.GetName()+"{"+strings.Join(labels, ",")+"}"] = m.GetGauge().GetValue()
		}
	}

	return out
}

func rectangular(n int) window.Window {
	w := make(window.Window, n)
	for i := range w {
		w[i] = 1
	}

	return w
}

// A scalar and a spectrum whose names sanitise to the same metric name fail the scrape instead of panicking.
func TestCollectorConflictingNames(t *testing.T) {
	reader := metric.NewManualReader()

	spectrum, err := metric.NewSpectrum("audio.level", metric.PeakHold, metric.WithFrameSize(8), metric.WithHop(8))
	require.NoError(t, err)

	level, err := metric.NewInstrument("audio_level", func(x []float64, _ int) float64 { return feature.Peak(x) },
		metric.Max)
	require.NoError(t, err)

	m, err := metric.NewMeter(reader, spectrum, level)
	require.NoError(t, err)

	f := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: f, Samples: audiotest.BinCosine(8, 1, 0.5)}))

	reg := prometheus.NewRegistry()
	require.NoError(t, reg.Register(promexp.NewCollector(reader)))

	_, err = reg.Gather()
	require.Error(t, err)
}
