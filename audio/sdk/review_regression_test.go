package sdk_test

import (
	"context"
	"io"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/internal/testutil"
	"github.com/zalgonoise/x/audio/sdk"
)

// Regression tests for the 2026-10-27 review of the sdk package (CLAUDE.md §10).

// formatReader is a SampleReader with an arbitrary format that delivers silence until n samples were read.
type formatReader struct {
	f audio.Format
	n int
}

func (r *formatReader) Format() audio.Format { return r.f }

func (r *formatReader) ReadSamples(dst []float64) (int, error) {
	if r.n <= 0 {
		return 0, io.EOF
	}

	n := min(len(dst), r.n)
	clear(dst[:n])
	r.n -= n

	return n, nil
}

func withFormat(f audio.Format, n int) sdk.Option {
	return sdk.WithDecoder(func(io.Reader) (audio.SampleReader, error) { return &formatReader{f: f, n: n}, nil })
}

func TestPipelineRejectsHostileFormats(t *testing.T) {
	huge := audio.Format{SampleRate: 8000, Channels: 1 << 46, BitDepth: 16, Encoding: audio.EncodingPCM}
	zeroRate := audio.Format{SampleRate: 0, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}

	p, err := sdk.NewPipeline(bytesSource(nil), []sdk.Exporter{&recorder{}}, withFormat(huge, 1))
	require.NoError(t, err)
	require.ErrorIs(t, p.Run(t.Context()), audio.ErrUnsupportedFormat, "no allocation from 2^46 channels")

	rec := &recorder{}
	p, err = sdk.NewPipeline(bytesSource(nil), []sdk.Exporter{rec}, withFormat(zeroRate, 1000))
	require.NoError(t, err)
	require.ErrorIs(t, p.Run(t.Context()), audio.ErrInvalidFormat)
	require.Empty(t, rec.buffers())
}

func TestPipelineRunWithCancelledContextExportsNothing(t *testing.T) {
	file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16}.EncodeInts([][]int32{{1, 2, 3}})
	rec := &recorder{}

	p, err := sdk.NewPipeline(bytesSource(file), []sdk.Exporter{rec})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	require.ErrorIs(t, p.Run(ctx), context.Canceled)
	require.Empty(t, rec.buffers())
}

// A Shutdown whose ctx ends while Run is stuck still shuts the exporters down once Run returns.
func TestPipelineShutdownCompletesAfterDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		opened := make(chan struct{})
		src := sdk.SourceFunc(func(context.Context) (io.ReadCloser, error) {
			close(opened)
			<-release // an Open that ignores its ctx, like os.Open on a FIFO

			return nil, io.ErrUnexpectedEOF
		})

		rec := &recorder{}

		p, err := sdk.NewPipeline(src, []sdk.Exporter{rec})
		require.NoError(t, err)

		ran := make(chan error, 1)

		go func() { ran <- p.Run(context.Background()) }()

		<-opened

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		require.ErrorIs(t, p.Shutdown(ctx), context.DeadlineExceeded)

		_, _, shutdowns := rec.counts()
		require.Zero(t, shutdowns, "Run has not returned yet")

		close(release)
		require.ErrorIs(t, <-ran, sdk.ErrShutdown)
		synctest.Wait()

		_, _, shutdowns = rec.counts()
		require.Equal(t, 1, shutdowns, "the exporters are shut down once Run returns")

		done, cancelDone := context.WithCancel(t.Context())
		cancelDone()

		for range 50 {
			require.NoError(t, p.Shutdown(done), "a known result wins over a done ctx")
		}
	})
}

func TestQueuedExporterShutdownReleasesBlockedProducers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		stuck := make(chan struct{})
		rec := &recorder{gate: stuck}

		q, err := sdk.NewQueuedExporter(rec, sdk.WithQueueSize(1), sdk.WithPolicy(sdk.Block))
		require.NoError(t, err)

		require.NoError(t, q.Export(t.Context(), buf(0))) // taken by the worker, stuck in the exporter
		synctest.Wait()
		require.NoError(t, q.Export(t.Context(), buf(1))) // queued

		blocked := make(chan error, 1)

		go func() { blocked <- q.Export(context.Background(), buf(2)) }()

		synctest.Wait()

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		go func() { _ = q.Shutdown(ctx) }() //nolint:errcheck // asserted through the producer

		require.ErrorIs(t, <-blocked, sdk.ErrShutdown, "a producer waiting for room gives up at Shutdown")

		close(stuck)
		synctest.Wait()
	})
}

func TestQueuedExporterDropsAbandonedFlushes(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		gate := make(chan struct{})
		rec := &recorder{gate: gate}

		q, err := sdk.NewQueuedExporter(rec)
		require.NoError(t, err)
		require.NoError(t, q.Export(t.Context(), buf(0)))
		synctest.Wait()

		for range 100 {
			ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond)
			require.ErrorIs(t, q.ForceFlush(ctx), context.DeadlineExceeded)
			cancel()
		}

		close(gate)
		require.NoError(t, q.ForceFlush(t.Context()))

		_, flushes, _ := rec.counts()
		require.Equal(t, 1, flushes, "abandoned flush requests never reach the exporter")
		require.NoError(t, q.Shutdown(t.Context()))
	})
}
