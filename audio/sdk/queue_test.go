package sdk_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/sdk"
)

var mono = audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}

// buf returns a one-sample Buffer whose sample and Frame are both i, to identify it.
func buf(i int) audio.Buffer {
	return audio.Buffer{Format: mono, Samples: []float64{float64(i)}, Frame: int64(i)}
}

func frames(bufs []audio.Buffer) []int64 {
	out := make([]int64, len(bufs))
	for i, b := range bufs {
		out[i] = b.Frame
	}

	return out
}

func TestQueuedExporterDropOldest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := &recorder{gate: make(chan struct{})}

		q, err := sdk.NewQueuedExporter(rec, sdk.WithQueueSize(2))
		require.NoError(t, err)

		require.NoError(t, q.Export(t.Context(), buf(0)))
		synctest.Wait() // the worker took buffer 0 and waits on the gate

		for i := 1; i <= 4; i++ {
			require.NoError(t, q.Export(t.Context(), buf(i)), "the producer never waits")
		}

		require.Equal(t, uint64(2), q.Dropped(), "1 and 2 made room for 3 and 4")

		close(rec.gate)
		require.NoError(t, q.Shutdown(t.Context()))
		require.Equal(t, []int64{0, 3, 4}, frames(rec.buffers()), "the exporter caught up with the latest audio")
	})
}

func TestQueuedExporterBlock(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := &recorder{gate: make(chan struct{})}

		q, err := sdk.NewQueuedExporter(rec, sdk.WithQueueSize(1), sdk.WithPolicy(sdk.Block))
		require.NoError(t, err)

		require.NoError(t, q.Export(t.Context(), buf(0)))
		synctest.Wait()
		require.NoError(t, q.Export(t.Context(), buf(1)), "queued")

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		require.ErrorIs(t, q.Export(ctx, buf(2)), context.DeadlineExceeded, "full: waits until ctx ends")

		done := make(chan error, 1)

		go func() { done <- q.Export(t.Context(), buf(3)) }()

		synctest.Wait()
		require.Empty(t, done, "still waiting for room")

		rec.gate <- struct{}{} // buffer 0 exported, buffer 1 dequeued: room for 3

		require.NoError(t, <-done)

		close(rec.gate)
		require.NoError(t, q.Shutdown(t.Context()))
		require.Equal(t, []int64{0, 1, 3}, frames(rec.buffers()), "nothing lost but the timed-out one")
		require.Zero(t, q.Dropped())
	})
}

func TestQueuedExporterFlushAndShutdown(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := &recorder{}

		q, err := sdk.NewQueuedExporter(rec, sdk.WithPolicy(sdk.Block))
		require.NoError(t, err)

		for i := range 5 {
			require.NoError(t, q.Export(t.Context(), buf(i)))
		}

		require.NoError(t, q.ForceFlush(t.Context()))

		exports, flushes, _ := rec.counts()
		require.Equal(t, 5, exports, "everything queued before ForceFlush was exported first")
		require.Equal(t, 1, flushes)

		require.NoError(t, q.Export(t.Context(), buf(5)))
		require.NoError(t, q.Shutdown(t.Context()), "drains, then shuts the exporter down")
		require.NoError(t, q.Shutdown(t.Context()), "idempotent")

		exports, _, shutdowns := rec.counts()
		require.Equal(t, 6, exports)
		require.Equal(t, 1, shutdowns)

		require.ErrorIs(t, q.Export(t.Context(), buf(6)), sdk.ErrShutdown)
		require.ErrorIs(t, q.ForceFlush(t.Context()), sdk.ErrShutdown)
	})
}

func TestQueuedExporterShutdownDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := &recorder{gate: make(chan struct{})} // never opened: the exporter is stuck

		q, err := sdk.NewQueuedExporter(rec, sdk.WithQueueSize(4))
		require.NoError(t, err)

		for i := range 3 {
			require.NoError(t, q.Export(t.Context(), buf(i)))
		}

		flushed := make(chan error, 1)

		go func() { flushed <- q.ForceFlush(context.Background()) }()

		synctest.Wait()

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		require.ErrorIs(t, q.Shutdown(ctx), context.DeadlineExceeded)
		require.ErrorIs(t, <-flushed, sdk.ErrShutdown, "a pending flush is released")

		synctest.Wait()

		_, _, shutdowns := rec.counts()
		require.Equal(t, 1, shutdowns, "the exporter is shut down even when the drain gave up")
		require.Equal(t, uint64(2), q.Dropped(), "buffers 1 and 2 were abandoned; 0 was cancelled in Export")
		require.Empty(t, rec.buffers())
	})
}

func TestQueuedExporterCopiesBuffers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := &recorder{gate: make(chan struct{})}

		q, err := sdk.NewQueuedExporter(rec)
		require.NoError(t, err)

		b := buf(7)
		require.NoError(t, q.Export(t.Context(), b))

		b.Samples[0] = -1 // the caller reuses its buffer

		close(rec.gate)
		require.NoError(t, q.Shutdown(t.Context()))
		require.Equal(t, []float64{7}, rec.samples())
	})
}

func TestQueuedExporterErrors(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		boom := errors.New("boom")

		var (
			mu   sync.Mutex
			errs []error
		)

		q, err := sdk.NewQueuedExporter(&recorder{exportErr: boom}, sdk.WithQueueErrorHandler(func(err error) {
			mu.Lock()
			defer mu.Unlock()

			errs = append(errs, err)
		}))
		require.NoError(t, err)
		require.NoError(t, q.Export(t.Context(), buf(0)), "errors are asynchronous")
		require.NoError(t, q.Shutdown(t.Context()))
		require.Len(t, errs, 1)
		require.ErrorIs(t, errs[0], boom)
	})

	_, err := sdk.NewQueuedExporter(nil)
	require.ErrorIs(t, err, sdk.ErrNilExporter)

	for name, opt := range map[string]sdk.QueueOption{
		"Size":         sdk.WithQueueSize(0),
		"Policy":       sdk.WithPolicy(sdk.Policy(9)),
		"ErrorHandler": sdk.WithQueueErrorHandler(nil),
	} {
		_, err := sdk.NewQueuedExporter(&recorder{}, opt)
		require.ErrorIs(t, err, sdk.ErrInvalidOption, name)
	}

	require.Equal(t, "drop-oldest", sdk.DropOldest.String())
	require.Equal(t, "block", sdk.Block.String())
}
