package poller_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/memory"
	"github.com/zalgonoise/x/gacha/poller"
	"github.com/zalgonoise/x/gacha/writer"
)

type order struct {
	ID    int    `json:"id"`
	Total string `json:"total"`
}

// TestEndToEnd writes orders concurrently through a batch writer that notifies the poller, which handles them with a
// flaky handler: every order fails twice before succeeding, and one order is invalid.
func TestEndToEnd(t *testing.T) {
	const orders, invalidID = 50, 13

	synctest.Test(t, func(t *testing.T) {
		repo := memory.New[[]byte]()
		notifier := poller.NewNotifier()

		w := writer.NewBatch(repo, gacha.JSON[order](), writer.WithNotify(notifier.Notify))

		var (
			mu       sync.Mutex
			attempts = make(map[int]int, orders)
			handled  = make(map[int]bool, orders)
		)

		p := poller.New(repo, gacha.JSON[order](), func(_ context.Context, o order) error {
			mu.Lock()
			defer mu.Unlock()

			if o.ID == invalidID {
				return gacha.Permanent(errBoom)
			}

			attempts[o.ID]++
			if attempts[o.ID] < 3 {
				return errBoom
			}

			handled[o.ID] = true

			return nil
		},
			poller.WithTrigger(poller.Any(notifier, poller.Ticker(time.Second))),
			poller.WithBatchSize(8),
			poller.WithConcurrency(4),
			poller.WithRetryBackoff(poller.Exponential(100*time.Millisecond, time.Second)),
		)

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)

		go func() {
			done <- p.Run(ctx)
		}()

		var wg sync.WaitGroup

		for id := range orders {
			wg.Go(func() {
				if err := w.Write(t.Context(), order{ID: id, Total: strconv.Itoa(id * 100)}); err != nil {
					t.Error(err)
				}
			})
		}

		wg.Wait()
		require.NoError(t, w.Close(t.Context()))

		// Leave time for two retries per order.
		time.Sleep(10 * time.Second)

		cancel()
		require.NoError(t, <-done)

		mu.Lock()
		require.Len(t, handled, orders-1)
		require.False(t, handled[invalidID])
		mu.Unlock()

		requireStats(t, repo, gacha.Stats{Completed: orders - 1, DeadLetters: 1})
	})
}

func BenchmarkPoller_Poll(b *testing.B) {
	for _, size := range []int{1, 100} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			repo := memory.New[int](memory.WithDeleteOnComplete())
			items := make([]int, size)

			p := poller.New(repo, gacha.Identity[int](), func(context.Context, int) error {
				return nil
			}, poller.WithBatchSize(size), poller.WithHeartbeat(0))

			b.ReportAllocs()

			for b.Loop() {
				if err := repo.Write(b.Context(), items...); err != nil {
					b.Fatal(err)
				}

				if _, err := p.Poll(b.Context()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
