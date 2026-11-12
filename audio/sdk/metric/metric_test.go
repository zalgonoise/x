package metric_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/dsp/feature"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/sdk"
	"github.com/zalgonoise/x/audio/sdk/metric"
)

var stereo = audio.Format{SampleRate: 1000, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM}

// constant returns a stereo Buffer of n frames holding l on the left channel and r on the right.
func constant(frame int64, n int, l, r float64) audio.Buffer {
	s := make([]float64, 0, 2*n)
	for range n {
		s = append(s, l, r)
	}

	return audio.Buffer{Format: stereo, Samples: s, Frame: frame}
}

// byKey indexes a batch by name and channel.
func byKey(batch []metric.Measurement) map[string]map[int]metric.Measurement {
	out := map[string]map[int]metric.Measurement{}

	for _, m := range batch {
		if out[m.Name] == nil {
			out[m.Name] = map[int]metric.Measurement{}
		}

		out[m.Name][m.Channel] = m
	}

	return out
}

func TestMeterPerChannel(t *testing.T) {
	r := metric.NewManualReader()

	m, err := metric.NewMeter(r, metric.Levels()...)
	require.NoError(t, err)

	require.NoError(t, m.Export(t.Context(), constant(0, 10, 0.5, -0.25)))
	require.NoError(t, m.Export(t.Context(), constant(10, 10, 1, -0.5)))

	batch, err := r.Collect(t.Context())
	require.NoError(t, err)
	require.Len(t, batch, 12, "6 instruments × 2 channels")

	got := byKey(batch)
	require.InDelta(t, 1, got["audio.peak"][0].Value, 0, "max of the two buffers' peaks")
	require.InDelta(t, 0.5, got["audio.peak"][1].Value, 0)
	require.InDelta(t, 0.75, got["audio.rms"][0].Value, 1e-12, "mean of 0.5 and 1")
	require.InDelta(t, -0.375, got["audio.dc_offset"][1].Value, 1e-12)
	require.InDelta(t, 1, got["audio.crest_factor"][0].Value, 1e-12)
	require.InDelta(t, 10, got["audio.clipped_samples"][0].Value, 0, "10 samples at full scale")
	require.Zero(t, got["audio.clipped_samples"][1].Value)
	require.Zero(t, got["audio.zero_crossing_rate"][0].Value)

	peak := got["audio.peak"][0]
	require.Equal(t, int64(0), peak.Start)
	require.Equal(t, int64(20), peak.End)
	require.Equal(t, 1000, peak.SampleRate)

	// delta: the next collection only covers what was recorded since
	require.NoError(t, m.Export(t.Context(), constant(20, 5, 0.1, 0.1)))

	batch, err = r.Collect(t.Context())
	require.NoError(t, err)

	got = byKey(batch)
	require.InDelta(t, 0.1, got["audio.peak"][0].Value, 1e-12)
	require.Equal(t, int64(20), got["audio.peak"][0].Start)
	require.Equal(t, int64(25), got["audio.peak"][0].End)

	batch, err = r.Collect(t.Context())
	require.NoError(t, err)
	require.Empty(t, batch, "nothing recorded since")

	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: stereo}), "an empty buffer is ignored")
}

func TestMeterMixdown(t *testing.T) {
	r := metric.NewManualReader()

	inst, err := metric.NewInstrument("dc", func(x []float64, _ int) float64 { return feature.DCOffset(x) }, metric.Last,
		metric.OnMixdown())
	require.NoError(t, err)
	require.Equal(t, "dc", inst.Name())

	m, err := metric.NewMeter(r, inst)
	require.NoError(t, err)
	require.NoError(t, m.Export(t.Context(), constant(0, 4, 0.5, -0.25)))

	batch, err := r.Collect(t.Context())
	require.NoError(t, err)
	require.Len(t, batch, 1)
	require.Equal(t, metric.Mixdown, batch[0].Channel)
	require.InDelta(t, 0.125, batch[0].Value, 1e-12)
}

func TestAggregations(t *testing.T) {
	for _, tc := range []struct {
		name string
		agg  func() metric.Aggregation
		want float64
	}{
		{"Last", metric.Last, 2},
		{"Max", metric.Max, 3},
		{"Min", metric.Min, -1},
		{"Mean", metric.Mean, 4.0 / 3},
		{"Sum", metric.Sum, 4},
	} {
		a := tc.agg()

		_, ok := a.Collect()
		require.False(t, ok, "%s: nothing recorded", tc.name)

		for _, v := range []float64{3, -1, 2} {
			a.Record(v, 1)
		}

		v, ok := a.Collect()
		require.True(t, ok)
		require.InDelta(t, tc.want, v, 1e-12, tc.name)

		_, ok = a.Collect()
		require.False(t, ok, "%s: Collect resets", tc.name)
	}

	m := metric.Max()
	m.Record(1, 1)
	m.Record(math.NaN(), 1)
	m.Record(2, 1)

	weighted := metric.Mean()
	weighted.Record(0.5, 4410)
	weighted.Record(-0.5, 1)

	v, _ := weighted.Collect()
	require.InDelta(t, (0.5*4410-0.5)/4411, v, 1e-12, "a 1-frame tail weighs one frame")

	v, _ = m.Collect()
	require.True(t, math.IsNaN(v), "a NaN measurement is not hidden")
}

// recordingEmitter keeps every batch it receives.
type recordingEmitter struct {
	mu        sync.Mutex
	batches   [][]metric.Measurement
	flushes   int
	shutdowns int
	block     chan struct{} // when set, Emit waits on it or ctx
}

func (e *recordingEmitter) Emit(ctx context.Context, batch []metric.Measurement) error {
	if e.block != nil {
		select {
		case <-e.block:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.batches = append(e.batches, append([]metric.Measurement(nil), batch...))

	return nil
}

func (e *recordingEmitter) ForceFlush(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.flushes++

	return nil
}

func (e *recordingEmitter) Shutdown(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.shutdowns++

	return nil
}

func (e *recordingEmitter) state() (batches, flushes, shutdowns int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return len(e.batches), e.flushes, e.shutdowns
}

func peakOnly(t *testing.T) metric.Instrument {
	t.Helper()

	inst, err := metric.NewInstrument("peak", func(x []float64, _ int) float64 { return feature.Peak(x) }, metric.Max)
	require.NoError(t, err)

	return inst
}

func TestPeriodicReader(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		em := &recordingEmitter{}

		r, err := metric.NewPeriodicReader(em, metric.WithInterval(time.Second))
		require.NoError(t, err)

		m, err := metric.NewMeter(r, peakOnly(t))
		require.NoError(t, err)

		require.NoError(t, m.Export(t.Context(), constant(0, 10, 0.5, 0.25)))
		time.Sleep(time.Second)
		synctest.Wait()

		batches, _, _ := em.state()
		require.Equal(t, 1, batches, "one interval, one batch")

		time.Sleep(3 * time.Second)
		synctest.Wait()

		batches, _, _ = em.state()
		require.Equal(t, 1, batches, "empty intervals emit nothing")

		require.NoError(t, m.Export(t.Context(), constant(10, 10, 0.1, 0.1)))
		require.NoError(t, m.ForceFlush(t.Context()))

		batches, flushes, _ := em.state()
		require.Equal(t, 2, batches, "ForceFlush emits now")
		require.Equal(t, 1, flushes)

		require.NoError(t, m.Export(t.Context(), constant(20, 10, 0.2, 0.2)))
		require.NoError(t, m.Shutdown(t.Context()))
		require.NoError(t, m.Shutdown(t.Context()), "idempotent")

		batches, _, shutdowns := em.state()
		require.Equal(t, 3, batches, "Shutdown emits the last batch")
		require.Equal(t, 1, shutdowns)
		require.InDelta(t, 0.2, em.batches[2][0].Value, 1e-12)

		require.ErrorIs(t, m.Export(t.Context(), constant(30, 1, 0, 0)), sdk.ErrShutdown)
		require.ErrorIs(t, m.ForceFlush(t.Context()), sdk.ErrShutdown)
		require.ErrorIs(t, r.ForceFlush(t.Context()), sdk.ErrShutdown)
	})
}

func TestPeriodicReaderShutdownDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		em := &recordingEmitter{block: make(chan struct{})} // never released

		var (
			mu   sync.Mutex
			errs []error
		)

		r, err := metric.NewPeriodicReader(em, metric.WithInterval(time.Second), metric.WithEmitTimeout(time.Second),
			metric.WithErrorHandler(func(err error) {
				mu.Lock()
				defer mu.Unlock()

				errs = append(errs, err)
			}))
		require.NoError(t, err)

		m, err := metric.NewMeter(r, peakOnly(t))
		require.NoError(t, err)
		require.NoError(t, m.Export(t.Context(), constant(0, 10, 0.5, 0.25)))

		time.Sleep(1500 * time.Millisecond) // the periodic emit is stuck until its timeout
		synctest.Wait()

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		require.ErrorIs(t, m.Shutdown(ctx), context.DeadlineExceeded)

		time.Sleep(time.Second) // the stuck periodic emit times out; the goroutine then sees quit and exits
		synctest.Wait()

		mu.Lock()
		require.NotEmpty(t, errs, "the timed-out periodic emit was reported")
		mu.Unlock()

		_, _, shutdowns := em.state()
		require.Equal(t, 1, shutdowns, "the emitter is shut down anyway")
	})
}

func TestReaderRules(t *testing.T) {
	defer goleak.VerifyNone(t)

	r := metric.NewManualReader()

	_, err := r.Collect(t.Context())
	require.ErrorIs(t, err, metric.ErrNoMeter)

	_, err = metric.NewMeter(r, peakOnly(t))
	require.NoError(t, err)

	_, err = metric.NewMeter(r, peakOnly(t))
	require.ErrorIs(t, err, metric.ErrRegistered, "one meter per reader")

	require.NoError(t, r.ForceFlush(t.Context()))
	require.NoError(t, r.Shutdown(t.Context()))

	_, err = r.Collect(t.Context())
	require.ErrorIs(t, err, sdk.ErrShutdown)

	// a periodic reader that never served a meter still shuts its emitter down
	em := &recordingEmitter{}
	pr, err := metric.NewPeriodicReader(em)
	require.NoError(t, err)
	require.NoError(t, pr.ForceFlush(t.Context()))
	require.NoError(t, pr.Shutdown(t.Context()))

	_, flushes, shutdowns := em.state()
	require.Equal(t, 1, flushes)
	require.Equal(t, 1, shutdowns)

	_, err = metric.NewMeter(pr, peakOnly(t))
	require.ErrorIs(t, err, sdk.ErrShutdown)
}

func TestValidation(t *testing.T) {
	_, err := metric.NewInstrument("", func([]float64, int) float64 { return 0 }, metric.Max)
	require.ErrorIs(t, err, metric.ErrInvalidInstrument)

	_, err = metric.NewInstrument("x", nil, metric.Max)
	require.ErrorIs(t, err, metric.ErrInvalidInstrument)

	_, err = metric.NewInstrument("x", func([]float64, int) float64 { return 0 }, nil)
	require.ErrorIs(t, err, metric.ErrInvalidInstrument)

	_, err = metric.NewMeter(nil)
	require.ErrorIs(t, err, sdk.ErrInvalidOption)

	_, err = metric.NewMeter(metric.NewManualReader(), metric.Instrument{})
	require.ErrorIs(t, err, metric.ErrInvalidInstrument)

	_, err = metric.NewPeriodicReader(nil)
	require.ErrorIs(t, err, sdk.ErrInvalidOption)

	for name, opt := range map[string]metric.PeriodicOption{
		"Interval":     metric.WithInterval(0),
		"EmitTimeout":  metric.WithEmitTimeout(-time.Second),
		"ErrorHandler": metric.WithErrorHandler(nil),
	} {
		_, err := metric.NewPeriodicReader(&recordingEmitter{}, opt)
		require.ErrorIs(t, err, sdk.ErrInvalidOption, name)
	}

	_, err = metric.NewSlogEmitter(nil, slog.LevelInfo)
	require.ErrorIs(t, err, sdk.ErrInvalidOption)
}

// The meter in a pipeline: per-channel peaks of a fixture match the peaks of its decoded channels.
func TestMeterInPipeline(t *testing.T) {
	defer goleak.VerifyNone(t)

	file := fixtures.WAV(t, "amen_kick/amen_kick_stereo_16bit_44100hz.wav")

	d, err := wav.NewDecoder(bytes.NewReader(file))
	require.NoError(t, err)

	want, err := audio.ReadAll(d)
	require.NoError(t, err)

	r := metric.NewManualReader()

	m, err := metric.NewMeter(r, peakOnly(t))
	require.NoError(t, err)

	p, err := sdk.NewPipeline(sdk.SourceFunc(func(context.Context) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(file)), nil
	}), []sdk.Exporter{m})
	require.NoError(t, err)
	require.NoError(t, p.Run(t.Context()))

	batch, err := r.Collect(t.Context())
	require.NoError(t, err)

	got := byKey(batch)
	for ch := range 2 {
		require.InDelta(t, feature.Peak(want.Channel(ch, nil)), got["peak"][ch].Value, 0, "channel %d", ch)
		require.Equal(t, int64(want.Frames()), got["peak"][ch].End)
	}

	require.NoError(t, p.Shutdown(t.Context()))
}

func TestSlogEmitter(t *testing.T) {
	var out bytes.Buffer

	em, err := metric.NewSlogEmitter(slog.New(slog.NewJSONHandler(&out, nil)), slog.LevelInfo)
	require.NoError(t, err)

	require.NoError(t, em.Emit(t.Context(), []metric.Measurement{
		{Name: "audio.peak", Channel: 1, Start: 10, End: 20, SampleRate: 8000, Value: 0.5},
	}))

	var rec map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &rec))
	require.Equal(t, "audio.peak", rec["msg"])
	require.InDelta(t, 1, rec["channel"], 0)
	require.InDelta(t, 0.5, rec["value"], 0)
	require.InDelta(t, 20, rec["end_frame"], 0)

	require.NoError(t, em.ForceFlush(t.Context()))
	require.NoError(t, em.Shutdown(t.Context()))
	require.ErrorIs(t, em.Emit(t.Context(), nil), sdk.ErrShutdown)

	var called bool

	f := metric.EmitFunc(func(context.Context, []metric.Measurement) error {
		called = true

		return nil
	})
	require.NoError(t, f.Emit(t.Context(), nil))
	require.NoError(t, f.ForceFlush(t.Context()))
	require.NoError(t, f.Shutdown(t.Context()))
	require.True(t, called)
}

// Review round 3 (2026-10-27).

func TestMeterWindows(t *testing.T) {
	r := metric.NewManualReader()

	m, err := metric.NewMeter(r, peakOnly(t))
	require.NoError(t, err)

	require.NoError(t, m.Export(t.Context(), constant(44000, 100, 0.5, 0.5)))
	require.NoError(t, m.Export(t.Context(), constant(0, 10, 0.25, 0.25)), "a reconnect restarts at frame 0")

	other := stereo
	other.SampleRate = 8
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: other, Samples: []float64{0.1, 0.1}, Frame: 10}))
	require.NoError(t, m.Export(t.Context(), constant(500, 10, 0.75, 0.75)), "a gap extends the window")

	batch, err := r.Collect(t.Context())
	require.NoError(t, err)

	var left []metric.Measurement

	for _, x := range batch {
		if x.Channel == 0 {
			left = append(left, x)
		}
	}

	require.Equal(t, []metric.Measurement{
		{Name: "peak", Channel: 0, Start: 44000, End: 44100, SampleRate: 1000, Value: 0.5},
		{Name: "peak", Channel: 0, Start: 0, End: 10, SampleRate: 1000, Value: 0.25},
		{Name: "peak", Channel: 0, Start: 10, End: 11, SampleRate: 8, Value: 0.1},
		{Name: "peak", Channel: 0, Start: 500, End: 510, SampleRate: 1000, Value: 0.75},
	}, left, "every window is one coherent stretch of one stream")
}

func TestMeterCollectOrder(t *testing.T) {
	r := metric.NewManualReader()

	m, err := metric.NewMeter(r, metric.Levels()[:2]...) // peak, rms
	require.NoError(t, err)

	mono := audio.Format{SampleRate: 1000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}
	require.NoError(t, m.Export(t.Context(), audio.Buffer{Format: mono, Samples: []float64{0.5}}))
	require.NoError(t, m.Export(t.Context(), constant(1, 1, 0.5, 0.5)))

	batch, err := r.Collect(t.Context())
	require.NoError(t, err)

	order := make([]string, 0, len(batch))
	for _, x := range batch {
		order = append(order, fmt.Sprintf("%s/%d", x.Name, x.Channel))
	}

	require.Equal(t, []string{"audio.peak/0", "audio.peak/1", "audio.rms/0", "audio.rms/1"}, order,
		"instrument, then channel, whatever order the channels appeared in (channel 0 continues one window)")
}

func TestPeriodicReaderFlushDeadlineAndPanics(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		em := &recordingEmitter{block: make(chan struct{})}

		r, err := metric.NewPeriodicReader(em, metric.WithInterval(time.Hour))
		require.NoError(t, err)

		m, err := metric.NewMeter(r, peakOnly(t))
		require.NoError(t, err)
		require.NoError(t, m.Export(t.Context(), constant(0, 10, 0.5, 0.25)))

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		require.ErrorIs(t, m.ForceFlush(ctx), context.DeadlineExceeded, "the flush's emit is bounded by its ctx")

		close(em.block)
		require.NoError(t, m.Shutdown(t.Context()))
	})

	synctest.Test(t, func(t *testing.T) {
		r, err := metric.NewPeriodicReader(metric.EmitFunc(func(context.Context, []metric.Measurement) error {
			panic("emitter bug")
		}), metric.WithInterval(time.Hour))
		require.NoError(t, err)

		m, err := metric.NewMeter(r, peakOnly(t))
		require.NoError(t, err)
		require.NoError(t, m.Export(t.Context(), constant(0, 10, 0.5, 0.25)))
		require.ErrorIs(t, m.ForceFlush(t.Context()), metric.ErrEmitterPanic, "the process survives")
		require.NoError(t, m.Shutdown(t.Context()))
	})
}

func TestConcurrentShutdownHonoursOwnContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		em := &recordingEmitter{block: make(chan struct{})}

		r, err := metric.NewPeriodicReader(em, metric.WithInterval(time.Hour))
		require.NoError(t, err)

		m, err := metric.NewMeter(r, peakOnly(t))
		require.NoError(t, err)
		require.NoError(t, m.Export(t.Context(), constant(0, 10, 0.5, 0.25)))

		first := make(chan error, 1)

		go func() { first <- m.Shutdown(context.Background()) }() // the final emit blocks

		synctest.Wait()

		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()

		require.ErrorIs(t, m.Shutdown(ctx), context.DeadlineExceeded, "a second caller waits only up to its own ctx")

		close(em.block)
		require.NoError(t, <-first)
		require.NoError(t, m.Shutdown(ctx), "the result, once known, wins over a done ctx")
	})
}
