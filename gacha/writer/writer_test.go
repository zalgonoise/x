package writer_test

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/memory"
	"github.com/zalgonoise/x/gacha/writer"
)

// Compile-time checks that the writers implement gacha.Writer.
var (
	_ gacha.Writer[int] = (*writer.Sync[int, string])(nil)
	_ gacha.Writer[int] = (*writer.Batch[int, string])(nil)
)

var errBoom = errors.New("boom")

// recorder is a gacha.Sink that records its calls. Each call waits for a token on release (when set), then returns the
// next error of errs (nil once they run out).
type recorder struct {
	mu      sync.Mutex
	calls   [][]string
	errs    []error
	release chan struct{}
}

func (r *recorder) Write(ctx context.Context, data ...string) error {
	if r.release != nil {
		select {
		case <-r.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.calls = append(r.calls, slices.Clone(data))

	if len(r.errs) == 0 {
		return nil
	}

	err := r.errs[0]
	r.errs = r.errs[1:]

	return err
}

func (r *recorder) recorded() [][]string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return slices.Clone(r.calls)
}

// itoa marshals ints into strings, failing on negative values.
var itoa = gacha.MarshalerFunc[int, string](func(item int) (string, error) {
	if item < 0 {
		return "", errBoom
	}

	return strconv.Itoa(item), nil
})

type counter struct {
	mu sync.Mutex
	n  int
}

func (c *counter) inc() {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
}

func (c *counter) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.n
}

func TestSync_Write(t *testing.T) {
	for _, testcase := range []struct {
		name          string
		items         []int
		maxBatchSize  int
		retries       int
		sinkErrs      []error
		wantsCalls    [][]string
		wantsErr      error
		wantsNotified int
	}{
		{
			name:          "Simple",
			items:         []int{1, 2, 3},
			wantsCalls:    [][]string{{"1", "2", "3"}},
			wantsNotified: 1,
		},
		{
			name:  "Empty",
			items: nil,
		},
		{
			name:          "Chunked",
			items:         []int{1, 2, 3, 4, 5},
			maxBatchSize:  2,
			wantsCalls:    [][]string{{"1", "2"}, {"3", "4"}, {"5"}},
			wantsNotified: 1,
		},
		{
			name:     "MarshalError",
			items:    []int{1, -1, 3},
			wantsErr: writer.ErrMarshal,
		},
		{
			name:       "SinkError",
			items:      []int{1},
			sinkErrs:   []error{errBoom},
			wantsCalls: [][]string{{"1"}},
			wantsErr:   errBoom,
		},
		{
			name:          "ChunkFailureKeepsEarlierChunks",
			items:         []int{1, 2, 3},
			maxBatchSize:  2,
			sinkErrs:      []error{nil, errBoom},
			wantsCalls:    [][]string{{"1", "2"}, {"3"}},
			wantsErr:      errBoom,
			wantsNotified: 1,
		},
		{
			name:          "Retried",
			items:         []int{1},
			retries:       2,
			sinkErrs:      []error{errBoom, errBoom},
			wantsCalls:    [][]string{{"1"}, {"1"}, {"1"}},
			wantsNotified: 1,
		},
		{
			name:       "RetriesExhausted",
			items:      []int{1},
			retries:    1,
			sinkErrs:   []error{errBoom, errBoom},
			wantsCalls: [][]string{{"1"}, {"1"}},
			wantsErr:   errBoom,
		},
		{
			name:       "PermanentNotRetried",
			items:      []int{1},
			retries:    3,
			sinkErrs:   []error{gacha.Permanent(errBoom)},
			wantsCalls: [][]string{{"1"}},
			wantsErr:   errBoom,
		},
		{
			name:       "ContextErrorNotRetried",
			items:      []int{1},
			retries:    3,
			sinkErrs:   []error{context.DeadlineExceeded},
			wantsCalls: [][]string{{"1"}},
			wantsErr:   context.DeadlineExceeded,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			sink := &recorder{errs: testcase.sinkErrs}
			notified := &counter{}

			w := writer.NewSync(sink, itoa,
				writer.WithMaxBatchSize(testcase.maxBatchSize),
				writer.WithRetry(testcase.retries, nil),
				writer.WithNotify(notified.inc),
			)

			err := w.Write(t.Context(), testcase.items...)
			if testcase.wantsErr != nil {
				require.ErrorIs(t, err, testcase.wantsErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testcase.wantsCalls, sink.recorded())
			require.Equal(t, testcase.wantsNotified, notified.count())
		})
	}
}

func TestSync_RetryBackoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &recorder{errs: []error{errBoom, errBoom}}
		w := writer.NewSync(sink, itoa, writer.WithRetry(3, func(retry int) time.Duration {
			return time.Duration(retry) * time.Second
		}))

		start := time.Now()

		require.NoError(t, w.Write(t.Context(), 1))
		require.Equal(t, 3*time.Second, time.Since(start))
		require.Len(t, sink.recorded(), 3)
	})
}

func TestSync_RetryCanceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &recorder{errs: []error{errBoom, errBoom, errBoom}}
		w := writer.NewSync(sink, itoa, writer.WithRetry(3, func(int) time.Duration {
			return time.Second
		}))

		ctx, cancel := context.WithTimeout(t.Context(), 1500*time.Millisecond)
		defer cancel()

		err := w.Write(ctx, 1)
		require.ErrorIs(t, err, errBoom)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Len(t, sink.recorded(), 2)
	})
}

func TestSync_Concurrent(t *testing.T) {
	const writers, items = 16, 50

	sink := &recorder{}
	w := writer.NewSync(sink, itoa)

	var wg sync.WaitGroup

	for i := range writers {
		wg.Go(func() {
			for j := range items {
				if err := w.Write(t.Context(), i*items+j); err != nil {
					t.Error(err)
				}
			}
		})
	}

	wg.Wait()
	require.Len(t, sink.recorded(), writers*items)
}

func TestSync_Repository(t *testing.T) {
	type order struct {
		ID    int    `json:"id"`
		Total string `json:"total"`
	}

	repo := memory.New[[]byte]()
	orders := []order{{ID: 1, Total: "9.99"}, {ID: 2, Total: "0.50"}}

	// T = order and R = []byte are inferred from the arguments.
	require.NoError(t, writer.NewSync(repo, gacha.JSON[order]()).Write(t.Context(), orders...))

	entries, err := repo.Claim(t.Context(), 10, time.Minute)
	require.NoError(t, err)
	require.Len(t, entries, len(orders))

	for i := range entries {
		got, err := gacha.JSON[order]().Unmarshal(entries[i].Data)
		require.NoError(t, err)
		require.Equal(t, orders[i], got)
	}
}
