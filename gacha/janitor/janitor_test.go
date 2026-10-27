package janitor_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/janitor"
	"github.com/zalgonoise/x/gacha/memory"
	"github.com/zalgonoise/x/gacha/poller"
)

var errBoom = errors.New("boom")

// settleAll claims every claimable entry of repo and settles it with status.
func settleAll(t *testing.T, repo *memory.Repository[string], status gacha.Status) {
	t.Helper()

	entries, err := repo.Claim(t.Context(), 100, time.Minute)
	require.NoError(t, err)

	outcomes := make([]gacha.Outcome, len(entries))

	for i := range entries {
		outcomes[i] = gacha.Outcome{ID: entries[i].ID, Attempts: entries[i].Attempts, Status: status}
	}

	require.NoError(t, repo.Settle(t.Context(), outcomes...))
}

func requireStats(t *testing.T, repo *memory.Repository[string], wants gacha.Stats) {
	t.Helper()

	stats, err := repo.Stats(t.Context())
	require.NoError(t, err)
	require.Equal(t, wants, stats)
}

// archiver records archived values, failing with the configured errors first.
type archiver struct {
	mu       sync.Mutex
	archived []string
	errs     map[string][]error
}

func (a *archiver) archive(_ context.Context, data string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if errs := a.errs[data]; len(errs) > 0 {
		a.errs[data] = errs[1:]

		return errs[0]
	}

	a.archived = append(a.archived, data)

	return nil
}

func TestJanitor_Completed(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := memory.New[string]()
		require.NoError(t, repo.Write(t.Context(), "a", "b", "c"))
		settleAll(t, repo, gacha.StatusCompleted)

		a := &archiver{}
		j := janitor.New(repo.Completed(time.Hour), a.archive, janitor.WithInterval(10*time.Minute))

		res, err := j.Poll(t.Context())
		require.NoError(t, err)
		require.Zero(t, res.Claimed, "entries are kept for the retention period")

		time.Sleep(time.Hour)

		res, err = j.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 3, Completed: 3}, res)
		require.Equal(t, []string{"a", "b", "c"}, a.archived)
		requireStats(t, repo, gacha.Stats{})
	})
}

func TestJanitor_ArchiveFailures(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := memory.New[string]()
		require.NoError(t, repo.Write(t.Context(), "ok", "flaky", "broken"))
		settleAll(t, repo, gacha.StatusCompleted)

		a := &archiver{errs: map[string][]error{
			"flaky":  {errBoom},
			"broken": {gacha.Permanent(errBoom)},
		}}
		j := janitor.New(repo.Completed(0), a.archive,
			janitor.WithInterval(10*time.Minute),
			janitor.WithLease(time.Hour),
		)

		res, err := j.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 3, Completed: 1, Retried: 1, Failed: 1}, res)
		requireStats(t, repo, gacha.Stats{Completed: 2})

		// The failed archive is retried at the next run. The permanent failure waits for its lease to expire.
		time.Sleep(10 * time.Minute)

		res, err = j.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 1, Completed: 1}, res)
		requireStats(t, repo, gacha.Stats{Completed: 1})

		time.Sleep(time.Hour)

		res, err = j.Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 1, Completed: 1}, res)
		require.Equal(t, []string{"ok", "flaky", "broken"}, a.archived)
		requireStats(t, repo, gacha.Stats{})
	})
}

func TestJanitor_DeadLetters(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := memory.New[string]()
		require.NoError(t, repo.Write(t.Context(), "poison"))

		// Claim the entry more times than a regular poller allows, as a crash loop would.
		for range 12 {
			_, err := repo.Claim(t.Context(), 1, time.Second)
			require.NoError(t, err)

			time.Sleep(time.Second)
		}

		settleAll(t, repo, gacha.StatusFailed)
		requireStats(t, repo, gacha.Stats{DeadLetters: 1})

		// A nil archive deletes without archiving. The entry's 13 attempts do not stop the janitor.
		res, err := janitor.New(repo.DeadLetters(0), nil).Poll(t.Context())
		require.NoError(t, err)
		require.Equal(t, poller.Result{Claimed: 1, Completed: 1}, res)
		requireStats(t, repo, gacha.Stats{})
	})
}

func TestJanitor_Run(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := memory.New[string]()

		var (
			mu      sync.Mutex
			results []poller.Result
		)

		j := janitor.New(repo.Completed(0), nil,
			janitor.WithInterval(time.Minute),
			janitor.WithBatchSize(2),
			janitor.WithPollerOptions(poller.WithObserver(func(res poller.Result) {
				mu.Lock()
				results = append(results, res)
				mu.Unlock()
			})),
		)

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)

		go func() {
			done <- j.Run(ctx)
		}()

		synctest.Wait()

		require.NoError(t, repo.Write(t.Context(), "a", "b", "c"))
		settleAll(t, repo, gacha.StatusCompleted)

		time.Sleep(time.Minute)
		synctest.Wait()

		mu.Lock()
		require.Equal(t, []poller.Result{
			{},
			{Claimed: 2, Completed: 2},
			{Claimed: 1, Completed: 1},
		}, results, "each run drains the view in batches")
		mu.Unlock()

		requireStats(t, repo, gacha.Stats{})

		cancel()
		require.NoError(t, <-done)
	})
}
