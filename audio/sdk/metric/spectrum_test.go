package metric_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"runtime"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/dsp/spectral"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/internal/testutil"
	"github.com/zalgonoise/x/audio/sdk"
	"github.com/zalgonoise/x/audio/sdk/metric"
)

const (
	specRate = 8000
	specSize = 256 // bins every 31.25 Hz
	specHop  = 128
)

var specFormat = audio.Format{SampleRate: specRate, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM}

// twoTones returns n stereo frames: a sine of ampL at hzL on the left, and of ampR at hzR on the right.
func twoTones(n int, hzL, ampL, hzR, ampR float64) []float64 {
	return testutil.Interleave(testutil.Sine(n, specRate, hzL, ampL, 0), testutil.Sine(n, specRate, hzR, ampR, 0.4))
}

// exportChunks exports interleaved stereo samples in buffers of the given frame counts (the rest in one buffer).
func exportChunks(t *testing.T, m *metric.Meter, f audio.Format, frame int64, samples []float64, chunks ...int) {
	t.Helper()

	ch := f.Channels

	for _, n := range append(chunks, len(samples)/ch) {
		n = min(n, len(samples)/ch)
		require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: f, Samples: samples[:n*ch], Frame: frame}))
		samples, frame = samples[n*ch:], frame+int64(n)
	}
}

func spectrumInstrument(t *testing.T, agg func() metric.SpectrumAggregation, opts ...metric.InstrumentOption,
) metric.Instrument {
	t.Helper()

	opts = append([]metric.InstrumentOption{metric.WithFrameSize(specSize), metric.WithHop(specHop)}, opts...)

	inst, err := metric.NewSpectrum("spectrum", agg, opts...)
	require.NoError(t, err)
	require.Equal(t, metric.KindSpectrum, inst.Kind())

	return inst
}

func collectOne(t *testing.T, r *metric.ManualReader) map[int]metric.Measurement {
	t.Helper()

	batch, err := r.Collect(t.Context())
	require.NoError(t, err)

	out := map[int]metric.Measurement{}

	for _, m := range batch {
		require.Equal(t, metric.KindSpectrum, m.Kind)
		require.NotContains(t, out, m.Channel, "one window per channel")

		out[m.Channel] = m
	}

	return out
}

func strongest(bins []metric.Bin) metric.Bin {
	best := bins[0]

	for _, b := range bins {
		if b.Value > best.Value {
			best = b
		}
	}

	return best
}

// Each channel's spectrum peaks on its own tone, at its amplitude, on float Hz bins.
func TestSpectrumPerChannel(t *testing.T) {
	for name, agg := range map[string]func() metric.SpectrumAggregation{
		"peak hold": metric.PeakHold,
		"mean":      metric.MeanSpectrum,
	} {
		t.Run(name, func(t *testing.T) {
			r := metric.NewManualReader()

			m, err := metric.NewMeter(r, spectrumInstrument(t, agg))
			require.NoError(t, err)

			// bin-centred tones: 1000 Hz is bin 32, 2031.25 Hz is bin 65
			exportChunks(t, m, specFormat, 0, twoTones(4000, 1000, 0.5, 2031.25, 0.25), 100, 333, 1, 1000)

			got := collectOne(t, r)
			require.Len(t, got, 2)

			for ch, want := range map[int]metric.Bin{0: {Frequency: 1000, Value: 0.5}, 1: {Frequency: 2031.25, Value: 0.25}} {
				ms := got[ch]
				require.Equal(t, "spectrum", ms.Name)
				require.Equal(t, int64(0), ms.Start)
				require.Equal(t, int64((4000-specSize)/specHop*specHop+specSize), ms.End, "the end of the last whole frame")
				require.Equal(t, specRate, ms.SampleRate)
				require.Len(t, ms.Spectrum, specSize/2+1, "every bin, 0 Hz to Nyquist")

				peak := strongest(ms.Spectrum)
				require.InDelta(t, want.Frequency, peak.Frequency, 0, "channel %d", ch)
				require.InDelta(t, want.Value, peak.Value, 1e-3, "channel %d", ch)
			}

			// the bins: centres k·31.25 Hz, edges half a bin either side, clipped to [0, Nyquist]
			bins := got[0].Spectrum
			require.Equal(t, metric.Bin{Frequency: 0, Low: 0, High: 15.625, Value: bins[0].Value}, bins[0])
			require.Equal(t, metric.Bin{Frequency: 31.25, Low: 15.625, High: 46.875, Value: bins[1].Value}, bins[1])
			require.Equal(t, metric.Bin{Frequency: 4000, Low: 3984.375, High: 4000, Value: bins[128].Value}, bins[128])
		})
	}
}

// The carry across buffers makes the spectra independent of how the stream is cut into buffers.
func TestSpectrumIndependentOfBuffering(t *testing.T) {
	signal := testutil.Interleave(testutil.Chirp(3000, specRate, 50, 3500, 0.9), testutil.Noise(3000, 0.3, 5))

	measure := func(chunks ...int) map[int]metric.Measurement {
		r := metric.NewManualReader()

		m, err := metric.NewMeter(r, spectrumInstrument(t, metric.MeanSpectrum))
		require.NoError(t, err)

		exportChunks(t, m, specFormat, 0, signal, chunks...)

		return collectOne(t, r)
	}

	want := measure()
	for _, chunks := range [][]int{{1}, {255, 1, 1}, {128, 128, 128}, {7, 300, 2000}} {
		got := measure(chunks...)
		for ch := range 2 {
			require.Len(t, got[ch].Spectrum, len(want[ch].Spectrum))

			for i := range want[ch].Spectrum {
				require.InDelta(t, want[ch].Spectrum[i].Value, got[ch].Spectrum[i].Value, 1e-12, "chunks %v", chunks)
			}
		}
	}
}

// counting records how many spectra it received, and their weights.
type counting struct {
	spectra, frames *int
	bins            int
}

func (c counting) Record(v []float64, frames int) {
	*c.spectra++
	*c.frames += frames
}

func (c counting) Collect() ([]float64, bool) { return make([]float64, c.bins), *c.spectra > 0 }

func countingInstrument(t *testing.T, spectra, frames *int, opts ...metric.InstrumentOption) metric.Instrument {
	t.Helper()

	return spectrumInstrument(t, func() metric.SpectrumAggregation {
		return counting{spectra: spectra, frames: frames, bins: specSize/2 + 1}
	}, append([]metric.InstrumentOption{metric.OnMixdown()}, opts...)...)
}

func TestSpectrumFrames(t *testing.T) {
	var spectra, frames int

	m, err := metric.NewMeter(metric.NewManualReader(), countingInstrument(t, &spectra, &frames))
	require.NoError(t, err)

	exportChunks(t, m, specFormat, 0, make([]float64, 2*1000), 300, 300)
	require.Equal(t, (1000-specSize)/specHop+1, spectra, "one spectrum per hop once a frame is full")
	require.Equal(t, spectra*specHop, frames, "each spectrum covers one hop")

	spectra = 0

	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: specFormat, Samples: make([]float64, 2*200), Frame: 0}))
	require.Zero(t, spectra, "a restart resets the carry: 200 frames do not complete a frame")

	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: specFormat, Samples: make([]float64, 2*200), Frame: 300}))
	require.Zero(t, spectra, "nor does a gap")

	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: specFormat, Samples: make([]float64, 2*200), Frame: 500}))
	require.Equal(t, 2, spectra, "contiguous buffers do: frames at 300 and 428")

	fast := specFormat
	fast.SampleRate = 2 * specRate

	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: fast, Samples: make([]float64, 2*200), Frame: 700}))
	require.Equal(t, 2, spectra, "nor does a new sample rate")

	none := specFormat
	none.SampleRate = 0

	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: none, Samples: make([]float64, 2*1000), Frame: 900}))
	require.Equal(t, 2, spectra, "buffers without a sample rate are ignored")
}

func TestSpectrumWindowsPerRate(t *testing.T) {
	r := metric.NewManualReader()

	m, err := metric.NewMeter(r, spectrumInstrument(t, metric.PeakHold, metric.OnMixdown()))
	require.NoError(t, err)

	tone := testutil.Interleave(testutil.Sine(1000, specRate, 1000, 0.5, 0), testutil.Sine(1000, specRate, 1000, 0.5, 0))
	exportChunks(t, m, specFormat, 0, tone)

	fast := specFormat
	fast.SampleRate = 2 * specRate
	exportChunks(t, m, fast, 1000, tone)

	batch, err := r.Collect(t.Context())
	require.NoError(t, err)
	require.Len(t, batch, 2, "a new rate opens a new window")

	require.Equal(t, metric.Mixdown, batch[0].Channel)
	require.Equal(t, specRate, batch[0].SampleRate)
	require.InDelta(t, 1000, strongest(batch[0].Spectrum).Frequency, 0)
	require.Equal(t, int64((1000-specSize)/specHop*specHop+specSize), batch[0].End, "the end of the last whole frame")

	// the same samples at twice the rate are a tone at twice the frequency
	require.Equal(t, 2*specRate, batch[1].SampleRate)
	require.InDelta(t, 2000, strongest(batch[1].Spectrum).Frequency, 0)
	require.InDelta(t, 62.5, batch[1].Spectrum[1].Frequency, 0)
	require.Equal(t, int64(1000), batch[1].Start, "the new rate starts a new frame")
	require.Equal(t, int64(1000+(1000-specSize)/specHop*specHop+specSize), batch[1].End)

	batch, err = r.Collect(t.Context())
	require.NoError(t, err)
	require.Empty(t, batch)

	exportChunks(t, m, specFormat, 0, tone[:2*100])

	batch, err = r.Collect(t.Context())
	require.NoError(t, err)
	require.Empty(t, batch, "a window in which no frame completed reports nothing")
}

// Bands reduce the spectrum to their levels: a tone reads its amplitude on its band.
func TestSpectrumBands(t *testing.T) {
	bands, err := spectral.OctaveBands(spectral.Octave, 31.5, 16000)
	require.NoError(t, err)

	r := metric.NewManualReader()

	m, err := metric.NewMeter(r, spectrumInstrument(t, metric.PeakHold, metric.WithBands(bands),
		metric.WithFrameSize(1024), metric.WithHop(1024)))
	require.NoError(t, err)

	// 1000 Hz on the left, 250 Hz on the right: both on bins (7.8125 Hz apart)
	exportChunks(t, m, specFormat, 0, twoTones(4096, 1000, 0.5, 250, 0.125))

	got := collectOne(t, r)

	// 31.5 Hz holds bins 3-5 at this resolution; 8 and 16 kHz lie above the 4 kHz Nyquist frequency
	centres := make([]float64, 0, len(got[0].Spectrum))
	for _, b := range got[0].Spectrum {
		centres = append(centres, b.Frequency)
	}

	require.Equal(t, []float64{31.5, 63, 125, 250, 500, 1000, 2000, 4000}, centres)
	require.Equal(t, bands[5].Low, got[0].Spectrum[5].Low)
	require.Equal(t, bands[5].High, got[0].Spectrum[5].High)

	for ch, want := range map[int]metric.Bin{0: {Frequency: 1000, Value: 0.5}, 1: {Frequency: 250, Value: 0.125}} {
		peak := strongest(got[ch].Spectrum)
		require.InDelta(t, want.Frequency, peak.Frequency, 0, "channel %d", ch)
		require.InEpsilon(t, want.Value, peak.Value, 0.01, "channel %d", ch)
	}

	// no band holds a bin when every band lies above the Nyquist frequency
	high, err := spectral.OctaveBands(spectral.Octave, 8000, 16000)
	require.NoError(t, err)

	r = metric.NewManualReader()

	m, err = metric.NewMeter(r, spectrumInstrument(t, metric.PeakHold, metric.WithBands(high)))
	require.NoError(t, err)

	exportChunks(t, m, specFormat, 0, twoTones(4096, 1000, 0.5, 250, 0.125))

	batch, err := r.Collect(t.Context())
	require.NoError(t, err)
	require.Empty(t, batch)
}

func TestSpectrumWindowOption(t *testing.T) {
	var calls []int

	rect := func(n int) window.Window {
		calls = append(calls, n)

		w := make(window.Window, n)
		for i := range w {
			w[i] = 1
		}

		return w
	}

	r := metric.NewManualReader()

	m, err := metric.NewMeter(r, spectrumInstrument(t, metric.PeakHold, metric.WithWindow(rect), metric.OnMixdown()))
	require.NoError(t, err)
	require.Equal(t, []int{specSize}, calls, "the window is built once")

	tone := testutil.Interleave(testutil.Sine(specSize, specRate, 1000, 0.5, 0), testutil.Sine(specSize, specRate, 1000,
		0.5, 0))
	exportChunks(t, m, specFormat, 0, tone)

	bins := collectOne(t, r)[metric.Mixdown].Spectrum
	for _, b := range bins {
		if b.Frequency != 1000 {
			require.InDelta(t, 0, b.Value, 1e-9, "a rectangular window does not leak a bin-centred tone (%g Hz)",
				b.Frequency)
		}
	}
}

func TestPeakHold(t *testing.T) {
	a := metric.PeakHold()

	_, ok := a.Collect()
	require.False(t, ok)

	in := []float64{1, 2, 0.5}
	a.Record(in, 10)
	in[0] = 100 // the aggregation does not keep the caller's slice

	a.Record([]float64{3, 1, 0.25}, 10)

	got, ok := a.Collect()
	require.True(t, ok)
	require.Equal(t, []float64{3, 2, 0.5}, got)

	_, ok = a.Collect()
	require.False(t, ok, "collect resets")

	a.Record([]float64{1, math.NaN()}, 1)
	a.Record([]float64{2, 5}, 1)

	got, _ = a.Collect()
	require.InDelta(t, 2, got[0], 0)
	require.True(t, math.IsNaN(got[1]), "NaN is sticky until the next collection")

	a.Record([]float64{1, 2}, 1)
	a.Record([]float64{7}, 1)

	got, _ = a.Collect()
	require.Equal(t, []float64{7}, got, "a spectrum of another length starts over")

	got[0] = 42 // the collected slice is the caller's

	a.Record([]float64{1}, 1)

	got, _ = a.Collect()
	require.Equal(t, []float64{1}, got)
}

func TestMeanSpectrum(t *testing.T) {
	a := metric.MeanSpectrum()

	_, ok := a.Collect()
	require.False(t, ok)

	a.Record([]float64{1, 0, 2}, 1)
	a.Record([]float64{3, 0, 2}, 1)

	got, ok := a.Collect()
	require.True(t, ok)
	require.InDeltaSlice(t, []float64{math.Sqrt(5), 0, 2}, got, 1e-12, "energy average")

	a.Record([]float64{1}, 1)
	a.Record([]float64{3}, 3)

	got, _ = a.Collect()
	require.InDelta(t, math.Sqrt(28.0/4), got[0], 1e-12, "weighted by frames")

	a.Record([]float64{-2}, 0)

	got, _ = a.Collect()
	require.InDelta(t, 2, got[0], 1e-12, "magnitudes; a zero-frame spectrum still counts once")

	a.Record([]float64{1e200, 3e-170}, 1)
	a.Record([]float64{1e200, 4e-170}, 1)

	got, _ = a.Collect()
	require.InEpsilon(t, 1e200, got[0], 1e-12, "no overflow")
	require.InEpsilon(t, math.Sqrt(12.5)*1e-170, got[1], 1e-12, "no underflow")

	a.Record([]float64{math.Inf(1), math.NaN(), 1}, 1)
	a.Record([]float64{1, 1, math.Inf(-1)}, 1)

	got, _ = a.Collect()
	require.True(t, math.IsInf(got[0], 1))
	require.True(t, math.IsNaN(got[1]))
	require.True(t, math.IsInf(got[2], 1))

	a.Record([]float64{1, 1}, 1)
	a.Record([]float64{4}, 1)

	got, _ = a.Collect()
	require.Equal(t, []float64{4}, got, "a spectrum of another length starts over")
}

func TestSpectrumValidation(t *testing.T) {
	bands, err := spectral.OctaveBands(spectral.Octave, 63, 8000)
	require.NoError(t, err)

	for name, opt := range map[string]metric.InstrumentOption{
		"frame size not a power of two": metric.WithFrameSize(1000),
		"frame size too small":          metric.WithFrameSize(2),
		"frame size too large":          metric.WithFrameSize(spectral.MaxSize * 2),
		"hop":                           metric.WithHop(0),
		"nil window":                    metric.WithWindow(nil),
		"no bands":                      metric.WithBands(nil),
		"overlapping bands":             metric.WithBands([]spectral.Band{bands[1], bands[0]}),
		"window length":                 metric.WithWindow(func(int) window.Window { return window.Hann(8) }),
		"zero window":                   metric.WithWindow(func(n int) window.Window { return make(window.Window, n) }),
	} {
		_, err := metric.NewSpectrum("s", metric.PeakHold, opt)
		require.ErrorIs(t, err, metric.ErrInvalidInstrument, name)
	}

	_, err = metric.NewSpectrum("", metric.PeakHold)
	require.ErrorIs(t, err, metric.ErrInvalidInstrument)

	_, err = metric.NewSpectrum("s", nil)
	require.ErrorIs(t, err, metric.ErrInvalidInstrument)

	_, err = metric.NewInstrument("peak", func([]float64, int) float64 { return 0 }, metric.Max, metric.WithBands(bands))
	require.ErrorIs(t, err, metric.ErrInvalidInstrument, "spectrum options on a scalar instrument fail")

	_, err = metric.NewMeter(metric.NewManualReader(), peakOnly(t), peakOnly(t))
	require.ErrorIs(t, err, metric.ErrInvalidInstrument, "duplicate names")

	_, err = metric.NewMeter(metric.NewManualReader(), metric.Instrument{})
	require.ErrorIs(t, err, metric.ErrInvalidInstrument, "the zero instrument")

	_, err = metric.Spectra(metric.WithHop(-1))
	require.ErrorIs(t, err, metric.ErrInvalidInstrument)

	require.Equal(t, metric.KindScalar, peakOnly(t).Kind())
	require.Equal(t, "scalar", metric.KindScalar.String())
	require.Equal(t, "spectrum", metric.KindSpectrum.String())
	require.Equal(t, "Kind(7)", metric.Kind(7).String())
}

// Spectra from a WAV stream through a pipeline, a periodic reader and the slog emitter: one JSON line per channel and
// instrument, with the band frequencies and their values.
func TestSpectraInPipeline(t *testing.T) {
	defer goleak.VerifyNone(t)

	synctest.Test(t, func(t *testing.T) {
		const rate = 44100

		// a bin-centred tone per channel at 4096 points: 1001.29 Hz (bin 93) on the left, 4005.18 Hz (bin 372) on the
		// right
		left := testutil.Sine(rate, rate, spectral.BinFrequency(93, 4096, rate), 0.5, 0)
		right := testutil.Sine(rate, rate, spectral.BinFrequency(372, 4096, rate), 0.25, 0)
		file := testutil.WAV{Format: testutil.FormatFloat, Channels: 2, SampleRate: rate, BitsPerSample: 32}.
			EncodeFloats([][]float64{left, right})

		bands, err := spectral.OctaveBands(spectral.ThirdOctave, 25, 20000)
		require.NoError(t, err)

		spectra, err := metric.Spectra(metric.WithBands(bands), metric.WithFrameSize(4096))
		require.NoError(t, err)

		var out bytes.Buffer

		em, err := metric.NewSlogEmitter(slog.New(slog.NewJSONHandler(&out, nil)), slog.LevelInfo)
		require.NoError(t, err)

		reader, err := metric.NewPeriodicReader(em, metric.WithInterval(time.Hour))
		require.NoError(t, err)

		m, err := metric.NewMeter(reader, spectra...)
		require.NoError(t, err)

		p, err := sdk.NewPipeline(sdk.SourceFunc(func(context.Context) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(file)), nil
		}), []sdk.Exporter{m})
		require.NoError(t, err)
		require.NoError(t, p.Run(t.Context()))
		require.Empty(t, out.String(), "nothing is emitted before the interval")

		require.NoError(t, p.Shutdown(t.Context()), "shutdown emits the last batch")

		type record struct {
			Msg         string    `json:"msg"`
			Channel     int       `json:"channel"`
			End         int64     `json:"end_frame"`
			Frequencies []float64 `json:"frequencies"`
			Values      []float64 `json:"values"`
		}

		var records []record

		sc := bufio.NewScanner(&out)
		for sc.Scan() {
			var r record
			require.NoError(t, json.Unmarshal(sc.Bytes(), &r))

			records = append(records, r)
		}

		require.Len(t, records, 4, "2 instruments × 2 channels")

		for _, r := range records {
			require.Contains(t, []string{"audio.spectrum.peak", "audio.spectrum.mean"}, r.Msg)
			require.Equal(t, int64((rate-4096)/2048*2048+4096), r.End, "the end of the last whole frame")
			require.Len(t, r.Values, len(r.Frequencies))
			require.Equal(t, 31.5, r.Frequencies[0], "25 Hz holds no bin at 10.8 Hz resolution")
			require.Equal(t, 20000.0, r.Frequencies[len(r.Frequencies)-1])

			want := map[int][2]float64{0: {1000, 0.5}, 1: {4000, 0.25}}[r.Channel]
			i := slices.Index(r.Values, slices.Max(r.Values))
			require.InDelta(t, want[0], r.Frequencies[i], 0, "%s channel %d", r.Msg, r.Channel)
			require.InEpsilon(t, want[1], r.Values[i], 0.01, "%s channel %d", r.Msg, r.Channel)
		}
	})
}

// A buffer declaring many channels costs little until each channel has carried a whole frame: series allocate their
// analysis buffers lazily and share their window and frequency layout.
func TestSpectrumManyChannelsAllocateLazily(t *testing.T) {
	const channels = 4096

	spectra, err := metric.Spectra()
	require.NoError(t, err)

	m, err := metric.NewMeter(metric.NewManualReader(), spectra...)
	require.NoError(t, err)

	f := audio.Format{SampleRate: 48000, Channels: channels, BitDepth: 16, Encoding: audio.EncodingPCM}
	b := audio.Buffer{Format: f, Samples: make([]float64, channels)}

	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	require.NoError(t, m.Export(t.Context(), b))

	runtime.ReadMemStats(&after)

	// 2 instruments × 4096 channels; preallocated 2048-point buffers, windows and layouts would take about 800 MB
	require.Less(t, after.TotalAlloc-before.TotalAlloc, uint64(16<<20))
}

// A spectrum's [Start, End) are the frames of the STFT frames it aggregates, even when a frame started before the
// previous collection.
func TestSpectrumCoversAnalysedFrames(t *testing.T) {
	r := metric.NewManualReader()

	inst := spectrumInstrument(t, metric.PeakHold, metric.WithHop(specSize), metric.WithWindow(rectangular),
		metric.OnMixdown())

	m, err := metric.NewMeter(r, inst)
	require.NoError(t, err)

	tone := testutil.Interleave(testutil.Sine(200, specRate, 1000, 0.5, 0), testutil.Sine(200, specRate, 1000, 0.5, 0))
	exportChunks(t, m, specFormat, 5000, tone)

	batch, err := r.Collect(t.Context())
	require.NoError(t, err)
	require.Empty(t, batch, "no frame completed")

	exportChunks(t, m, specFormat, 5200, make([]float64, 2*400))

	batch, err = r.Collect(t.Context())
	require.NoError(t, err)
	require.Len(t, batch, 1)
	require.Greater(t, strongest(batch[0].Spectrum).Value, 0.1, "the first frame holds the tone of frames [5000, 5200)")
	require.Equal(t, int64(5000), batch[0].Start, "the first frame started before the collection")
	require.Equal(t, int64(5000+2*specSize), batch[0].End, "two whole frames")

	exportChunks(t, m, specFormat, 5600, make([]float64, 2*200)) // completes the frame [5512, 5768) carried since 5512

	batch, err = r.Collect(t.Context())
	require.NoError(t, err)
	require.Len(t, batch, 1)
	require.Equal(t, int64(5000+2*specSize), batch[0].Start)
	require.Equal(t, int64(5000+3*specSize), batch[0].End)
}

func rectangular(n int) window.Window {
	w := make(window.Window, n)
	for i := range w {
		w[i] = 1
	}

	return w
}

// A built Instrument never changes: options only configure the Instrument being built.
func TestInstrumentsAreImmutable(t *testing.T) {
	a, err := metric.NewSpectrum("a", metric.PeakHold, metric.WithFrameSize(specSize))
	require.NoError(t, err)

	b, err := metric.NewSpectrum("b", metric.PeakHold, metric.WithFrameSize(specSize))
	require.NoError(t, err)

	r := metric.NewManualReader()

	m, err := metric.NewMeter(r, a, b)
	require.NoError(t, err)

	mono := audio.Format{SampleRate: specRate, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: mono, Samples: make([]float64, specSize)}))

	batch, err := r.Collect(t.Context())
	require.NoError(t, err)
	require.Len(t, batch, 2)

	for _, ms := range batch {
		require.Len(t, ms.Spectrum, specSize/2+1, ms.Name)
	}
}

// JSON has no NaN or infinities: the slog emitter writes them as strings, so every line keeps its shape.
func TestSlogEmitterNonFinite(t *testing.T) {
	var out bytes.Buffer

	em, err := metric.NewSlogEmitter(slog.New(slog.NewJSONHandler(&out, nil)), slog.LevelInfo)
	require.NoError(t, err)

	require.NoError(t, em.Emit(t.Context(), []metric.Measurement{
		{Name: "audio.peak", Value: math.NaN()},
		{Name: "audio.spectrum.peak", Kind: metric.KindSpectrum, Spectrum: []metric.Bin{
			{Frequency: 0, Value: 0.5}, {Frequency: 1000, Value: math.Inf(1)}, {Frequency: 2000, Value: math.Inf(-1)},
		}},
	}))

	sc := bufio.NewScanner(&out)

	var recs []map[string]any

	for sc.Scan() {
		var rec map[string]any
		require.NoError(t, json.Unmarshal(sc.Bytes(), &rec), sc.Text())

		recs = append(recs, rec)
	}

	require.Len(t, recs, 2)
	require.Equal(t, "NaN", recs[0]["value"])
	require.Equal(t, []any{0.5, "+Inf", "-Inf"}, recs[1]["values"])
	require.Equal(t, []any{0.0, 1000.0, 2000.0}, recs[1]["frequencies"])
}
