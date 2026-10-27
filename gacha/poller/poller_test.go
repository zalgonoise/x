package poller_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/memory"
	"github.com/zalgonoise/x/gacha/poller"
)

var errBoom = errors.New("boom")

func newRepo(t *testing.T, items ...int) *memory.Repository[int] {
	t.Helper()

	repo := memory.New[int]()
	require.NoError(t, repo.Write(t.Context(), items...))

	return repo
}

func requireStats(t *testing.T, repo gacha.Stater, wants gacha.Stats) {
	t.Helper()

	stats, err := repo.Stats(t.Context())
	require.NoError(t, err)
	require.Equal(t, wants, stats)
}

func TestPoller_Poll(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1, 2, 3)

		var (
			got []int
			ids []uuid.UUID
		)

		p := poller.New(repo, gacha.Identity[int](), func(ctx context.Context, item int) error {
			id, ok := gacha.IDFromContext(ctx)
			require.True(t, ok)

			got = append(got, item)
			ids = append(ids, id)

			return nil
		})

		res, err := p.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 3, Completed: 3}, res)
		require.Equal(t, []int{1, 2, 3}, got, "entries are handled in claim order")
		require.Len(t, ids, 3)
		require.NotEqual(t, ids[0], ids[1])
		requireStats(t, repo, gacha.Stats{Completed: 3})

		res, err = p.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{}, res)
	})
}

func TestPoller_Policy(t *testing.T) {
	failing := gacha.UnmarshalerFunc[int, int](func(int) (int, error) {
		return 0, errBoom
	})

	for _, testcase := range []struct {
		name        string
		handler     func(ctx context.Context, item int) error
		unmarshaler gacha.Unmarshaler[int, int]
		opts        []cfg.Option[poller.Config]
		// expiredClaims is the number of claims taken (and left to expire) before the poll.
		expiredClaims int
		wantsResult   poller.Result
		wantsStats    gacha.Stats
		wantsCalls    int
	}{
		{
			name:        "Success",
			handler:     func(context.Context, int) error { return nil },
			wantsResult: poller.Result{Claimed: 1, Completed: 1},
			wantsStats:  gacha.Stats{Completed: 1},
			wantsCalls:  1,
		},
		{
			name:        "ErrorRetries",
			handler:     func(context.Context, int) error { return errBoom },
			wantsResult: poller.Result{Claimed: 1, Retried: 1},
			wantsStats:  gacha.Stats{Pending: 1},
			wantsCalls:  1,
		},
		{
			name:        "PanicRetries",
			handler:     func(context.Context, int) error { panic("boom") },
			wantsResult: poller.Result{Claimed: 1, Retried: 1},
			wantsStats:  gacha.Stats{Pending: 1},
			wantsCalls:  1,
		},
		{
			name: "TimeoutRetries",
			handler: func(ctx context.Context, _ int) error {
				<-ctx.Done()

				return ctx.Err()
			},
			opts:        []cfg.Option[poller.Config]{poller.WithHandlerTimeout(time.Second)},
			wantsResult: poller.Result{Claimed: 1, Retried: 1},
			wantsStats:  gacha.Stats{Pending: 1},
			wantsCalls:  1,
		},
		{
			name:        "PermanentDeadLetters",
			handler:     func(context.Context, int) error { return gacha.Permanent(errBoom) },
			wantsResult: poller.Result{Claimed: 1, Failed: 1},
			wantsStats:  gacha.Stats{DeadLetters: 1},
			wantsCalls:  1,
		},
		{
			name:        "UnmarshalErrorDeadLetters",
			handler:     func(context.Context, int) error { return nil },
			unmarshaler: failing,
			wantsResult: poller.Result{Claimed: 1, Failed: 1},
			wantsStats:  gacha.Stats{DeadLetters: 1},
		},
		{
			name:    "UnmarshalPanicDeadLetters",
			handler: func(context.Context, int) error { return nil },
			unmarshaler: gacha.UnmarshalerFunc[int, int](func(int) (int, error) {
				panic("corrupt")
			}),
			wantsResult: poller.Result{Claimed: 1, Failed: 1},
			wantsStats:  gacha.Stats{DeadLetters: 1},
		},
		{
			name:        "LastAttemptDeadLetters",
			handler:     func(context.Context, int) error { return errBoom },
			opts:        []cfg.Option[poller.Config]{poller.WithMaxAttempts(1)},
			wantsResult: poller.Result{Claimed: 1, Failed: 1},
			wantsStats:  gacha.Stats{DeadLetters: 1},
			wantsCalls:  1,
		},
		{
			name:          "CrashLoopDeadLettersWithoutHandling",
			handler:       func(context.Context, int) error { return nil },
			opts:          []cfg.Option[poller.Config]{poller.WithMaxAttempts(2)},
			expiredClaims: 2,
			wantsResult:   poller.Result{Claimed: 1, Failed: 1},
			wantsStats:    gacha.Stats{DeadLetters: 1},
		},
		{
			name:          "UnlimitedAttempts",
			handler:       func(context.Context, int) error { return errBoom },
			opts:          []cfg.Option[poller.Config]{poller.WithMaxAttempts(0)},
			expiredClaims: 50,
			wantsResult:   poller.Result{Claimed: 1, Retried: 1},
			wantsStats:    gacha.Stats{Pending: 1},
			wantsCalls:    1,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				repo := newRepo(t, 42)

				for range testcase.expiredClaims {
					_, err := repo.Claim(t.Context(), 1, time.Second)
					require.NoError(t, err)

					time.Sleep(time.Second)
				}

				unmarshaler := testcase.unmarshaler
				if unmarshaler == nil {
					unmarshaler = gacha.Identity[int]()
				}

				var calls int

				p := poller.New(repo, unmarshaler, func(ctx context.Context, item int) error {
					calls++

					require.Equal(t, 42, item)

					return testcase.handler(ctx, item)
				}, testcase.opts...)

				res, err := p.Poll(t.Context())
				require.NoError(t, err)
				require.Equal(t, testcase.wantsResult, res)
				require.Equal(t, testcase.wantsCalls, calls)
				requireStats(t, repo, testcase.wantsStats)
			})
		})
	}
}

func TestPoller_RetryBackoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1)

		var retries []int

		p := poller.New(repo, gacha.Identity[int](), func(context.Context, int) error {
			return errBoom
		}, poller.WithMaxAttempts(3), poller.WithRetryBackoff(func(retry int) time.Duration {
			retries = append(retries, retry)

			return 5 * time.Second
		}))

		res, err := p.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, 1, res.Retried)

		res, err = p.Poll(t.Context())
		require.NoError(t, err)
		require.Zero(t, res.Claimed, "the entry waits for its retry backoff")

		for _, wants := range []poller.Result{{Claimed: 1, Retried: 1}, {Claimed: 1, Failed: 1}} {
			time.Sleep(5 * time.Second)

			res, err = p.Poll(t.Context())
			require.NoError(t, err)
			require.Equal(t, wants, res)
		}

		require.Equal(t, []int{1, 2}, retries)
		requireStats(t, repo, gacha.Stats{DeadLetters: 1})
	})
}

func TestPoller_Concurrency(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1, 2, 3, 4, 5, 6)

		var active, peak atomic.Int32

		p := poller.New(repo, gacha.Identity[int](), func(context.Context, int) error {
			n := active.Add(1)
			defer active.Add(-1)

			for {
				current := peak.Load()
				if n <= current || peak.CompareAndSwap(current, n) {
					break
				}
			}

			time.Sleep(time.Second)

			return nil
		}, poller.WithConcurrency(3))

		start := time.Now()

		res, err := p.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 6, Completed: 6}, res)
		require.Equal(t, int32(3), peak.Load())
		require.Equal(t, 2*time.Second, time.Since(start))
		requireStats(t, repo, gacha.Stats{Completed: 6})
	})
}

func TestPoller_SettleModes(t *testing.T) {
	for _, testcase := range []struct {
		name                string
		opts                []cfg.Option[poller.Config]
		wantsCompletedFirst int64
	}{
		{name: "PerItem", wantsCompletedFirst: 1},
		{name: "Batch", opts: []cfg.Option[poller.Config]{poller.WithBatchSettle()}, wantsCompletedFirst: 0},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				repo := newRepo(t, 1, 2)

				p := poller.New(repo, gacha.Identity[int](), func(ctx context.Context, item int) error {
					if item == 2 {
						stats, err := repo.Stats(ctx)
						require.NoError(t, err)
						require.Equal(t, testcase.wantsCompletedFirst, stats.Completed)
					}

					return nil
				}, testcase.opts...)

				res, err := p.Poll(t.Context())
				require.NoError(t, err)
				require.Equal(t, poller.Result{Claimed: 2, Completed: 2}, res)
				requireStats(t, repo, gacha.Stats{Completed: 2})
			})
		})
	}
}

func TestPoller_Heartbeat(t *testing.T) {
	for _, testcase := range []struct {
		name string
		opts []cfg.Option[poller.Config]
		// rivalClaims reports whether a rival claim takes the entry over while the handler runs.
		rivalClaims bool
		wantsStats  gacha.Stats
	}{
		{
			name:       "ExtendsLeases",
			wantsStats: gacha.Stats{Completed: 1},
		},
		{
			// An interval that is not shorter than the lease falls back to a third of it.
			name:       "IntervalClamped",
			opts:       []cfg.Option[poller.Config]{poller.WithHeartbeat(time.Hour)},
			wantsStats: gacha.Stats{Completed: 1},
		},
		{
			// Without heartbeats, the lease expires and a rival claims the entry. The poller's late outcome is fenced
			// out: the rival's claim stands.
			name:        "Disabled",
			opts:        []cfg.Option[poller.Config]{poller.WithHeartbeat(0)},
			rivalClaims: true,
			wantsStats:  gacha.Stats{Processing: 1},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				repo := newRepo(t, 1)
				opts := append([]cfg.Option[poller.Config]{poller.WithLease(3 * time.Second)}, testcase.opts...)

				p := poller.New(repo, gacha.Identity[int](), func(context.Context, int) error {
					time.Sleep(10 * time.Second)

					return nil
				}, opts...)

				rival := make(chan []gacha.Entry[int], 1)

				go func() {
					time.Sleep(5 * time.Second)

					entries, err := repo.Claim(t.Context(), 10, time.Minute)
					if err != nil {
						t.Error(err)
					}

					rival <- entries
				}()

				res, err := p.Poll(t.Context())
				require.NoError(t, err)
				require.Equal(t, poller.Result{Claimed: 1, Completed: 1}, res)
				require.Equal(t, testcase.rivalClaims, len(<-rival) == 1)
				requireStats(t, repo, testcase.wantsStats)
			})
		})
	}
}

// source wraps a gacha.Source with failing Claim or Settle calls.
type source struct {
	gacha.Source[int]

	claimErrs  atomic.Int32
	settleErrs atomic.Int32
}

func (s *source) Claim(ctx context.Context, limit int, lease time.Duration) ([]gacha.Entry[int], error) {
	if s.claimErrs.Add(-1) >= 0 {
		return nil, errBoom
	}

	return s.Source.Claim(ctx, limit, lease)
}

func (s *source) Settle(ctx context.Context, outcomes ...gacha.Outcome) error {
	if s.settleErrs.Add(-1) >= 0 {
		return errBoom
	}

	return s.Source.Settle(ctx, outcomes...)
}

func TestPoller_Errors(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1, 2)
		src := &source{Source: repo}
		src.claimErrs.Store(1)
		src.settleErrs.Store(1)

		p := poller.New[int, int](src, gacha.Identity[int](), func(context.Context, int) error {
			return nil
		}, poller.WithLease(time.Second))

		res, err := p.Poll(t.Context())
		require.ErrorIs(t, err, poller.ErrClaim)
		require.ErrorIs(t, err, errBoom)
		require.Equal(t, poller.Result{Err: err}, res)

		// A failed settle is retried with the rest of the batch.
		res, err = p.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 2, Completed: 2}, res)
		requireStats(t, repo, gacha.Stats{Completed: 2})

		// When the retry fails too, the poll reports it.
		require.NoError(t, repo.Write(t.Context(), 3))
		src.settleErrs.Store(2)

		res, err = p.Poll(t.Context())
		require.ErrorIs(t, err, poller.ErrSettle)
		require.ErrorIs(t, err, errBoom)
		require.Equal(t, 1, res.Completed, "the counts are the poller's verdicts")
		requireStats(t, repo, gacha.Stats{Processing: 1, Completed: 2})

		// The entry is claimed again once its lease expires.
		time.Sleep(time.Second)

		res, err = p.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 1, Completed: 1}, res)
		requireStats(t, repo, gacha.Stats{Completed: 3})
	})
}

func TestPoller_RunDrains(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1, 2, 3, 4, 5)

		var (
			mu      sync.Mutex
			results []poller.Result
		)

		p := poller.New(repo, gacha.Identity[int](), func(context.Context, int) error {
			return nil
		},
			poller.WithBatchSize(2),
			poller.WithTrigger(poller.Ticker(time.Hour)),
			poller.WithObserver(func(res poller.Result) {
				mu.Lock()
				results = append(results, res)
				mu.Unlock()
			}),
		)

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)

		go func() {
			done <- p.Run(ctx)
		}()

		synctest.Wait()

		mu.Lock()
		require.Equal(t, []poller.Result{
			{Claimed: 2, Completed: 2},
			{Claimed: 2, Completed: 2},
			{Claimed: 1, Completed: 1},
		}, results, "full batches are polled again right away")
		mu.Unlock()

		cancel()
		require.NoError(t, <-done)
		requireStats(t, repo, gacha.Stats{Completed: 5})
	})
}

func TestPoller_RunKeepsGoingOnErrors(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1)
		src := &source{Source: repo}
		src.claimErrs.Store(3)

		var failures atomic.Int32

		handled := make(chan struct{})

		p := poller.New[int, int](src, gacha.Identity[int](), func(context.Context, int) error {
			close(handled)

			return nil
		},
			poller.WithTrigger(poller.Ticker(time.Second)),
			poller.WithObserver(func(res poller.Result) {
				if res.Err != nil {
					failures.Add(1)
				}
			}),
		)

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)

		go func() {
			done <- p.Run(ctx)
		}()

		<-handled
		cancel()

		require.NoError(t, <-done)
		require.Equal(t, int32(3), failures.Load())
	})
}

func TestPoller_RunTriggerError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := poller.New(memory.New[int](), gacha.Identity[int](), func(context.Context, int) error {
			return nil
		}, poller.WithTrigger(poller.TriggerFunc(func(context.Context, poller.Result) error {
			return errBoom
		})))

		require.ErrorIs(t, p.Run(t.Context()), errBoom)
	})
}

func TestPoller_RunShutdown(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1, 2, 3)

		var calls atomic.Int32

		// With a single attempt, a regular failure would dead-letter the entry.
		p := poller.New(repo, gacha.Identity[int](), func(ctx context.Context, _ int) error {
			calls.Add(1)
			<-ctx.Done()

			return ctx.Err()
		}, poller.WithMaxAttempts(1))

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)

		go func() {
			done <- p.Run(ctx)
		}()

		synctest.Wait()
		cancel()

		require.NoError(t, <-done)
		require.Equal(t, int32(1), calls.Load(), "entries not started are not handled")
		requireStats(t, repo, gacha.Stats{Pending: 3})

		// Released entries are claimable right away, and the interrupted claim does not count: a poller allowing a
		// single attempt still handles them all.
		restarted := poller.New(repo, gacha.Identity[int](), func(context.Context, int) error {
			return nil
		}, poller.WithMaxAttempts(1))

		res, err := restarted.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 3, Completed: 3}, res)
	})
}

func TestPoller_Notifier(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := memory.New[int]()
		notifier := poller.NewNotifier()
		handled := make(chan int, 10)

		p := poller.New(repo, gacha.Identity[int](), func(_ context.Context, item int) error {
			handled <- item

			return nil
		}, poller.WithTrigger(notifier))

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)

		go func() {
			done <- p.Run(ctx)
		}()

		synctest.Wait()
		require.Empty(t, handled)

		require.NoError(t, repo.Write(t.Context(), 7))
		notifier.Notify()
		synctest.Wait()

		require.Equal(t, 7, <-handled)

		cancel()
		require.NoError(t, <-done)
	})
}

func TestPoller_WithoutDrain(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1, 2, 3, 4, 5)

		var polls atomic.Int32

		p := poller.New(repo, gacha.Identity[int](), func(context.Context, int) error {
			return nil
		},
			poller.WithBatchSize(2),
			poller.WithDrain(false),
			poller.WithTrigger(poller.Ticker(time.Hour)),
			poller.WithObserver(func(poller.Result) { polls.Add(1) }),
		)

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)

		go func() {
			done <- p.Run(ctx)
		}()

		synctest.Wait()
		require.Equal(t, int32(1), polls.Load(), "a full batch waits for the trigger")
		requireStats(t, repo, gacha.Stats{Pending: 3, Completed: 2})

		cancel()
		require.NoError(t, <-done)
	})
}

// blockingSettle is a gacha.Source whose Settle blocks until its context is done.
type blockingSettle struct {
	gacha.Source[int]
}

func (blockingSettle) Settle(ctx context.Context, _ ...gacha.Outcome) error {
	<-ctx.Done()

	return ctx.Err()
}

func TestPoller_SourceTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1)

		p := poller.New[int, int](blockingSettle{Source: repo}, gacha.Identity[int](), func(context.Context, int) error {
			return nil
		}, poller.WithSourceTimeout(3*time.Second))

		start := time.Now()

		_, err := p.Poll(t.Context())
		require.ErrorIs(t, err, poller.ErrSettle)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Equal(t, 6*time.Second, time.Since(start), "the settle and its retry with the batch each time out")
	})
}

func TestPoller_Logger(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1)
		src := &source{Source: repo}
		src.claimErrs.Store(1)

		var buf safeBuffer

		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		p := poller.New[int, int](src, gacha.Identity[int](), func(context.Context, int) error {
			panic("kaboom")
		}, poller.WithLogger(logger))

		_, err := p.Poll(t.Context())
		require.ErrorIs(t, err, poller.ErrClaim)

		_, err = p.Poll(t.Context())
		require.NoError(t, err)

		logs := buf.String()
		require.Contains(t, logs, "gacha: claim failed")
		require.Contains(t, logs, "gacha: handler panic")
		require.Contains(t, logs, "kaboom")
		require.Contains(t, logs, "gacha: poll")
	})
}

// safeBuffer is a bytes.Buffer safe for concurrent use.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

func TestPoller_UnmarshalPanicConcurrent(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1, 2, 3)

		p := poller.New(repo, gacha.UnmarshalerFunc[int, int](func(int) (int, error) {
			panic("corrupt")
		}), func(context.Context, int) error {
			return nil
		}, poller.WithConcurrency(3))

		res, err := p.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 3, Failed: 3}, res)
		requireStats(t, repo, gacha.Stats{DeadLetters: 3})
	})
}

// slowSettle is a gacha.Source whose Settle takes a while, keeping the gacha.Extender of the repository it wraps.
type slowSettle struct {
	*memory.Repository[int]

	delay time.Duration
}

func (s slowSettle) Settle(ctx context.Context, outcomes ...gacha.Outcome) error {
	time.Sleep(s.delay)

	return s.Repository.Settle(ctx, outcomes...)
}

func TestPoller_HeartbeatDuringBatchSettle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1)

		p := poller.New[int, int](slowSettle{Repository: repo, delay: 5 * time.Second}, gacha.Identity[int](),
			func(context.Context, int) error {
				return nil
			}, poller.WithLease(3*time.Second), poller.WithBatchSettle())

		rival := make(chan []gacha.Entry[int], 1)

		go func() {
			time.Sleep(4 * time.Second)

			entries, err := repo.Claim(t.Context(), 10, time.Minute)
			if err != nil {
				t.Error(err)
			}

			rival <- entries
		}()

		res, err := p.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 1, Completed: 1}, res)
		require.Empty(t, <-rival, "leases are renewed until the batch is settled")
		requireStats(t, repo, gacha.Stats{Completed: 1})
	})
}

// slowClaim is a gacha.Source whose Claim takes a while.
type slowClaim struct {
	*memory.Repository[int]

	delay time.Duration
}

func (s slowClaim) Claim(ctx context.Context, limit int, lease time.Duration) ([]gacha.Entry[int], error) {
	time.Sleep(s.delay)

	return s.Repository.Claim(ctx, limit, lease)
}

func TestPoller_ShutdownDuringClaim(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := newRepo(t, 1, 2)

		var calls atomic.Int32

		p := poller.New[int, int](slowClaim{Repository: repo, delay: time.Second}, gacha.Identity[int](),
			func(context.Context, int) error {
				calls.Add(1)

				return nil
			}, poller.WithMaxAttempts(1))

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)

		go func() {
			done <- p.Run(ctx)
		}()

		time.Sleep(500 * time.Millisecond)
		cancel()

		// The claim completes despite the shutdown, and its entries are released instead of stranded.
		require.NoError(t, <-done)
		require.Zero(t, calls.Load())
		requireStats(t, repo, gacha.Stats{Pending: 2})

		entries, err := repo.Claim(t.Context(), 10, time.Minute)
		require.NoError(t, err)
		require.Len(t, entries, 2)
		require.Equal(t, 1, entries[0].Attempts, "the interrupted claim does not count")
	})
}

func TestPoller_PollCanceled(t *testing.T) {
	repo := newRepo(t, 1)

	p := poller.New(repo, gacha.Identity[int](), func(context.Context, int) error {
		return nil
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := p.Poll(ctx)
	require.ErrorIs(t, err, context.Canceled)
	requireStats(t, repo, gacha.Stats{Pending: 1})
}
