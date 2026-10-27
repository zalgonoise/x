package writer_test

import (
	"context"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/writer"
)

// result is the outcome of an asynchronous Write.
type result struct {
	err     error
	elapsed time.Duration
}

// goWrite runs w.Write in a goroutine and returns a channel with its result.
func goWrite(ctx context.Context, w gacha.Writer[int], items ...int) <-chan result {
	out := make(chan result, 1)

	go func() {
		start := time.Now()
		err := w.Write(ctx, items...)
		out <- result{err: err, elapsed: time.Since(start)}
	}()

	return out
}

func closeBatch[T, R any](t *testing.T, b *writer.Batch[T, R]) {
	t.Helper()

	require.NoError(t, b.Close(context.Background()))
}

func TestBatch_Write(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &recorder{}
		notified := &counter{}
		b := writer.NewBatch(sink, itoa, writer.WithNotify(notified.inc))
		defer closeBatch(t, b)

		require.NoError(t, b.Write(t.Context()))
		require.NoError(t, b.Write(t.Context(), 1, 2, 3))
		require.Equal(t, [][]string{{"1", "2", "3"}}, sink.recorded())
		require.Equal(t, 1, notified.count())

		require.ErrorIs(t, b.Write(t.Context(), 4, -1), writer.ErrMarshal)
		require.Len(t, sink.recorded(), 1, "nothing is queued on marshal errors")
	})
}

func TestBatch_CoalescesWhileFlushing(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &recorder{release: make(chan struct{})}
		b := writer.NewBatch(sink, itoa)
		defer closeBatch(t, b)

		first := goWrite(t.Context(), b, 1)
		synctest.Wait() // The first flush is now blocked in the sink.

		rest := []<-chan result{
			goWrite(t.Context(), b, 2),
			goWrite(t.Context(), b, 3, 4),
			goWrite(t.Context(), b, 5),
		}
		synctest.Wait() // The other writers are now queued.

		close(sink.release)

		require.NoError(t, (<-first).err)

		for _, res := range rest {
			require.NoError(t, (<-res).err)
		}

		calls := sink.recorded()
		require.Len(t, calls, 2)
		require.Equal(t, []string{"1"}, calls[0])
		require.ElementsMatch(t, []string{"2", "3", "4", "5"}, calls[1])
	})
}

func TestBatch_Linger(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &recorder{}
		b := writer.NewBatch(sink, itoa, writer.WithLinger(10*time.Millisecond))
		defer closeBatch(t, b)

		first := goWrite(t.Context(), b, 1)

		time.Sleep(5 * time.Millisecond)

		second := goWrite(t.Context(), b, 2)

		res := <-first
		require.NoError(t, res.err)
		require.Equal(t, 10*time.Millisecond, res.elapsed, "the batch lingers from its first write")

		res = <-second
		require.NoError(t, res.err)
		require.Equal(t, 5*time.Millisecond, res.elapsed)

		calls := sink.recorded()
		require.Len(t, calls, 1)
		require.ElementsMatch(t, []string{"1", "2"}, calls[0])
	})
}

func TestBatch_MaxBatchSize(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &recorder{}
		b := writer.NewBatch(sink, itoa, writer.WithLinger(time.Hour), writer.WithMaxBatchSize(3))
		defer closeBatch(t, b)

		full := []<-chan result{goWrite(t.Context(), b, 1), goWrite(t.Context(), b, 2), goWrite(t.Context(), b, 3)}

		for _, res := range full {
			r := <-res
			require.NoError(t, r.err)
			require.Zero(t, r.elapsed, "a full batch flushes without lingering")
		}

		res := <-goWrite(t.Context(), b, 4)
		require.NoError(t, res.err)
		require.Equal(t, time.Hour, res.elapsed)

		// A single write larger than the batch size is flushed on its own, unsplit.
		res = <-goWrite(t.Context(), b, 5, 6, 7, 8)
		require.NoError(t, res.err)
		require.Zero(t, res.elapsed)

		calls := sink.recorded()
		require.Len(t, calls, 3)
		require.ElementsMatch(t, []string{"1", "2", "3"}, calls[0])
		require.Equal(t, []string{"4"}, calls[1])
		require.Equal(t, []string{"5", "6", "7", "8"}, calls[2])
	})
}

func TestBatch_Overflow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &recorder{}
		b := writer.NewBatch(sink, itoa, writer.WithLinger(time.Hour), writer.WithMaxBatchSize(3))
		defer closeBatch(t, b)

		first := goWrite(t.Context(), b, 1, 2)
		synctest.Wait() // The first batch is now lingering.

		second := goWrite(t.Context(), b, 3, 4)

		res := <-first
		require.NoError(t, res.err)
		require.Zero(t, res.elapsed, "a write that would overflow the batch flushes it")

		res = <-second
		require.NoError(t, res.err)
		require.Equal(t, time.Hour, res.elapsed, "the overflowing write starts the next batch")

		require.Equal(t, [][]string{{"1", "2"}, {"3", "4"}}, sink.recorded())
	})
}

func TestBatch_ErrorFanOut(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &recorder{release: make(chan struct{}), errs: []error{nil, errBoom}}
		notified := &counter{}
		b := writer.NewBatch(sink, itoa, writer.WithNotify(notified.inc))
		defer closeBatch(t, b)

		first := goWrite(t.Context(), b, 1)
		synctest.Wait()

		rest := []<-chan result{goWrite(t.Context(), b, 2), goWrite(t.Context(), b, 3)}
		synctest.Wait()

		close(sink.release)

		require.NoError(t, (<-first).err)

		for _, res := range rest {
			require.ErrorIs(t, (<-res).err, errBoom, "every caller of a batch gets its error")
		}

		require.Equal(t, 1, notified.count(), "failed flushes do not notify")
	})
}

func TestBatch_Retry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &recorder{errs: []error{errBoom}}
		b := writer.NewBatch(sink, itoa, writer.WithRetry(1, func(int) time.Duration { return time.Second }))
		defer closeBatch(t, b)

		res := <-goWrite(t.Context(), b, 1)
		require.NoError(t, res.err)
		require.Equal(t, time.Second, res.elapsed)
		require.Equal(t, [][]string{{"1"}, {"1"}}, sink.recorded())
	})
}

func TestBatch_FlushTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// The sink never releases: it returns when its context is done.
		sink := &recorder{release: make(chan struct{})}
		b := writer.NewBatch(sink, itoa, writer.WithFlushTimeout(time.Second))
		defer closeBatch(t, b)

		res := <-goWrite(t.Context(), b, 1)
		require.ErrorIs(t, res.err, context.DeadlineExceeded)
		require.Equal(t, time.Second, res.elapsed)
	})
}

func TestBatch_ContextDone(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &recorder{release: make(chan struct{})}
		b := writer.NewBatch(sink, itoa)
		defer closeBatch(t, b)

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		// The first write is queued, then times out while its flush is blocked: it is still persisted.
		flushing := goWrite(ctx, b, 1)
		synctest.Wait()

		// The second write times out before it is queued: it is not persisted.
		queued := goWrite(ctx, b, 2)

		for _, res := range []<-chan result{flushing, queued} {
			r := <-res
			require.ErrorIs(t, r.err, context.DeadlineExceeded)
			require.Equal(t, time.Second, r.elapsed)
		}

		close(sink.release)
		closeBatch(t, b)

		require.Equal(t, [][]string{{"1"}}, sink.recorded())
	})
}

func TestBatch_Close(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &recorder{release: make(chan struct{})}
		b := writer.NewBatch(sink, itoa, writer.WithLinger(time.Hour))

		pending := []<-chan result{goWrite(t.Context(), b, 1), goWrite(t.Context(), b, 2)}
		synctest.Wait()

		closed := make(chan error, 1)

		go func() {
			closed <- b.Close(t.Context())
		}()

		synctest.Wait()

		// Close stops lingering and waits for the pending writes to flush.
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		require.ErrorIs(t, b.Close(ctx), context.DeadlineExceeded)
		cancel()

		close(sink.release)

		for _, res := range pending {
			require.NoError(t, (<-res).err)
		}

		require.NoError(t, <-closed)
		require.NoError(t, b.Close(t.Context()), "Close is idempotent")
		require.ErrorIs(t, b.Write(t.Context(), 3), writer.ErrClosed)

		calls := sink.recorded()
		require.Len(t, calls, 1)
		require.ElementsMatch(t, []string{"1", "2"}, calls[0])
	})
}

func TestBatch_Concurrent(t *testing.T) {
	const writers, items = 16, 50

	sink := &recorder{}
	b := writer.NewBatch(sink, itoa, writer.WithMaxBatchSize(64))

	var wg sync.WaitGroup

	for i := range writers {
		wg.Go(func() {
			for j := range items {
				if err := b.Write(t.Context(), i*items+j); err != nil {
					t.Error(err)
				}
			}
		})
	}

	wg.Wait()
	closeBatch(t, b)

	var total int

	for _, call := range sink.recorded() {
		require.LessOrEqual(t, len(call), 64)

		total += len(call)
	}

	require.Equal(t, writers*items, total)
}
