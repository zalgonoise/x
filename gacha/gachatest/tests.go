package gachatest

import (
	"context"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/gacha"
)

func (s Suite[R]) testWriteEmpty(t *testing.T) {
	repo := s.New(t)

	require.NoError(t, repo.Write(t.Context()))
	require.Empty(t, claim(t, repo, 10, s.Lease))
	requireStats(t, repo, gacha.Stats{})
}

func (s Suite[R]) testCanceledContext(t *testing.T) {
	repo := s.New(t)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	require.ErrorIs(t, repo.Write(ctx, s.Value(0)), context.Canceled)

	_, err := repo.Claim(ctx, 1, s.Lease)
	require.ErrorIs(t, err, context.Canceled)

	err = repo.Settle(ctx, gacha.Outcome{ID: uuid.NewV7(), Attempts: 1, Status: gacha.StatusCompleted})
	require.ErrorIs(t, err, context.Canceled)

	require.Empty(t, claim(t, repo, 10, s.Lease))
}

func (s Suite[R]) testRoundTrip(t *testing.T) {
	repo := s.New(t)
	values := s.values(0, 5)
	write(t, repo, values...)

	entries := claim(t, repo, 10, s.Lease)
	require.Len(t, entries, len(values))

	seen := make(map[uuid.UUID]struct{}, len(entries))

	for i := range entries {
		require.Equal(t, values[i], entries[i].Data)
		require.Equal(t, 1, entries[i].Attempts)
		require.NotEqual(t, uuid.Nil(), entries[i].ID)
		require.False(t, entries[i].CreatedAt.IsZero())

		seen[entries[i].ID] = struct{}{}
	}

	require.Len(t, seen, len(entries), "IDs must be unique")

	// Leased entries are not claimed again.
	require.Empty(t, claim(t, repo, 10, s.Lease))
	requireStats(t, repo, gacha.Stats{Processing: 5})
}

func (s Suite[R]) testClaimOrder(t *testing.T) {
	repo := s.New(t)
	values := s.values(0, 6)
	write(t, repo, values[:1]...)
	write(t, repo, values[1:3]...)
	write(t, repo, values[3:]...)

	var got []R

	for {
		entries := claim(t, repo, 2, s.Lease)
		if len(entries) == 0 {
			break
		}

		require.LessOrEqual(t, len(entries), 2)

		got = append(got, data(entries)...)
	}

	require.Equal(t, values, got, "entries from one writer are claimed in write order")
}

func (s Suite[R]) testClaimLimits(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.values(0, 5)...)

	require.Len(t, claim(t, repo, 2, s.Lease), 2)
	require.Empty(t, claim(t, repo, 0, s.Lease))
	require.Empty(t, claim(t, repo, -1, s.Lease))
	require.Len(t, claim(t, repo, 10, s.Lease), 3)
	require.Empty(t, claim(t, repo, 10, s.Lease))
}

func (s Suite[R]) testClaimInvalidLease(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.Value(0))

	for _, lease := range []time.Duration{0, -time.Second} {
		_, err := repo.Claim(t.Context(), 1, lease)
		require.ErrorIs(t, err, gacha.ErrInvalidLease)
	}

	entries := claim(t, repo, 1, s.Lease)
	require.Len(t, entries, 1)
	require.Equal(t, 1, entries[0].Attempts, "invalid claims must not claim")
}

func (s Suite[R]) testLeaseExpiry(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.Value(0))

	first := claim(t, repo, 10, s.Lease)
	require.Len(t, first, 1)
	require.Equal(t, 1, first[0].Attempts)
	require.Empty(t, claim(t, repo, 10, s.Lease))

	s.expire()

	second := claim(t, repo, 10, s.Lease)
	require.Len(t, second, 1)
	require.Equal(t, first[0].ID, second[0].ID)
	require.Equal(t, first[0].Data, second[0].Data)
	require.Equal(t, 2, second[0].Attempts)
}

func (s Suite[R]) testConcurrentClaims(t *testing.T) {
	const (
		total   = 120
		workers = 8
		batch   = 7
	)

	repo := s.New(t)
	write(t, repo, s.values(0, total)...)

	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		seen = make(map[uuid.UUID]int, total)
		errs = make(chan error, workers)
	)

	for range workers {
		wg.Go(func() {
			for {
				entries, err := repo.Claim(t.Context(), batch, time.Minute)
				if err != nil {
					errs <- err

					return
				}

				if len(entries) == 0 {
					return
				}

				mu.Lock()
				for i := range entries {
					seen[entries[i].ID]++
				}
				mu.Unlock()

				if err = repo.Settle(t.Context(), completed(entries...)...); err != nil {
					errs <- err

					return
				}
			}
		})
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	require.Len(t, seen, total)

	for id, n := range seen {
		require.Equal(t, 1, n, "entry %s claimed %d times", id, n)
	}

	requireStats(t, repo, gacha.Stats{Completed: total})
}

func (s Suite[R]) testSettleEmpty(t *testing.T) {
	repo := s.New(t)

	require.NoError(t, repo.Settle(t.Context()))
}

func (s Suite[R]) testSettleCompleted(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.values(0, 2)...)

	entries := claim(t, repo, 10, s.Lease)
	settle(t, repo, completed(entries...)...)

	s.expire()

	require.Empty(t, claim(t, repo, 10, s.Lease), "completed entries are not claimed again")
	requireStats(t, repo, gacha.Stats{Completed: 2})
}

func (s Suite[R]) testSettleRetry(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.Value(0))

	entries := claim(t, repo, 1, s.Lease)
	settle(t, repo, retry(entries[0], 2*s.Lease, "retry me"))

	require.Empty(t, claim(t, repo, 1, s.Lease))
	requireStats(t, repo, gacha.Stats{Pending: 1})

	// The retry delay outlasts the claim's lease.
	time.Sleep(s.Lease)
	require.Empty(t, claim(t, repo, 1, s.Lease))

	s.expire()

	again := claim(t, repo, 1, s.Lease)
	require.Len(t, again, 1)
	require.Equal(t, entries[0].ID, again[0].ID)
	require.Equal(t, 2, again[0].Attempts)
}

func (s Suite[R]) testSettleRetryImmediately(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.Value(0))

	entries := claim(t, repo, 1, s.Lease)

	for attempt, delay := range []time.Duration{0, -time.Second} {
		settle(t, repo, retry(entries[0], delay, ""))

		entries = claim(t, repo, 1, s.Lease)
		require.Len(t, entries, 1)
		require.Equal(t, attempt+2, entries[0].Attempts)
	}
}

func (s Suite[R]) testSettleFailed(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.values(0, 2)...)

	entries := claim(t, repo, 10, s.Lease)
	settle(t, repo, failed(entries[0], "boom"), completed(entries[1])[0])

	s.expire()

	require.Empty(t, claim(t, repo, 10, s.Lease), "dead letters are not claimed again")
	requireStats(t, repo, gacha.Stats{Completed: 1, DeadLetters: 1})
}

func (s Suite[R]) testSettleMixed(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.values(0, 3)...)

	entries := claim(t, repo, 10, s.Lease)
	settle(t, repo,
		completed(entries[0])[0],
		retry(entries[1], 0, "again"),
		failed(entries[2], "boom"),
	)

	got := claim(t, repo, 10, s.Lease)
	require.Len(t, got, 1)
	require.Equal(t, entries[1].ID, got[0].ID)
	require.Equal(t, 2, got[0].Attempts)
	requireStats(t, repo, gacha.Stats{Processing: 1, Completed: 1, DeadLetters: 1})
}

func (s Suite[R]) testSettleFenced(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.Value(0))

	first := claim(t, repo, 1, s.Lease)

	s.expire()

	second := claim(t, repo, 1, s.Lease)
	require.Equal(t, 2, second[0].Attempts)

	// Outcomes of the expired claim are ignored.
	settle(t, repo, completed(first[0])...)
	settle(t, repo, failed(first[0], "stale"))
	requireStats(t, repo, gacha.Stats{Processing: 1})

	s.expire()

	third := claim(t, repo, 1, s.Lease)
	require.Len(t, third, 1)
	require.Equal(t, 3, third[0].Attempts, "stale outcomes must not settle the entry")

	settle(t, repo, completed(second[0])...)
	settle(t, repo, completed(third[0])...)

	s.expire()

	require.Empty(t, claim(t, repo, 1, s.Lease))
	requireStats(t, repo, gacha.Stats{Completed: 1})
}

func (s Suite[R]) testSettleReleased(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.Value(0))

	first := claim(t, repo, 1, s.Lease)
	require.Equal(t, 1, first[0].Attempts)

	settle(t, repo, released(first[0], "shutting down"))
	requireStats(t, repo, gacha.Stats{Pending: 1})

	// A released entry is claimable right away, and its claim does not count as an attempt.
	again := claim(t, repo, 1, s.Lease)
	require.Len(t, again, 1)
	require.Equal(t, first[0].ID, again[0].ID)
	require.Equal(t, 1, again[0].Attempts)

	s.expire()

	third := claim(t, repo, 1, s.Lease)
	require.Equal(t, 2, third[0].Attempts)

	// Releases are fenced like any other outcome.
	settle(t, repo, released(again[0], ""))
	requireStats(t, repo, gacha.Stats{Processing: 1})

	settle(t, repo, released(third[0], ""))

	fourth := claim(t, repo, 1, s.Lease)
	require.Len(t, fourth, 1)
	require.Equal(t, 2, fourth[0].Attempts)
}

func (s Suite[R]) testSettleIgnored(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.Value(0))

	entries := claim(t, repo, 1, s.Lease)

	// Unknown IDs are ignored.
	settle(t, repo, gacha.Outcome{ID: uuid.NewV7(), Attempts: 1, Status: gacha.StatusCompleted})
	settle(t, repo, completed(entries...)...)

	// Settled entries are ignored.
	settle(t, repo, retry(entries[0], 0, "late"))
	settle(t, repo, failed(entries[0], "late"))

	require.Empty(t, claim(t, repo, 1, s.Lease))
	requireStats(t, repo, gacha.Stats{Completed: 1})
}

func (s Suite[R]) testSettleInvalidStatus(t *testing.T) {
	repo := s.New(t)
	write(t, repo, s.values(0, 2)...)

	entries := claim(t, repo, 10, s.Lease)

	for _, status := range []gacha.Status{0, gacha.StatusProcessing, gacha.StatusReleased + 1} {
		err := repo.Settle(t.Context(),
			completed(entries[0])[0],
			gacha.Outcome{ID: entries[1].ID, Attempts: entries[1].Attempts, Status: status},
		)
		require.ErrorIs(t, err, gacha.ErrInvalidStatus, "status %s", status)
	}

	// Settle is atomic: the valid outcome was not applied either.
	requireStats(t, repo, gacha.Stats{Processing: 2})

	s.expire()

	require.Len(t, claim(t, repo, 10, s.Lease), 2)
}

func (s Suite[R]) testExtend(t *testing.T) {
	repo := s.New(t)

	extender, ok := repo.(gacha.Extender)
	if !ok {
		t.Skip("gacha.Extender not implemented")
	}

	write(t, repo, s.values(0, 2)...)
	entries := claim(t, repo, 10, s.Lease)

	require.ErrorIs(t, extender.Extend(t.Context(), 0, entries[0].ID), gacha.ErrInvalidLease)
	require.NoError(t, extender.Extend(t.Context(), 3*s.Lease, entries[0].ID, uuid.NewV7()))

	s.expire()

	// Only the entry that was not extended is claimable again.
	got := claim(t, repo, 10, s.Lease)
	require.Len(t, got, 1)
	require.Equal(t, entries[1].ID, got[0].ID)
	settle(t, repo, completed(got...)...)

	time.Sleep(2 * s.Lease)

	got = claim(t, repo, 10, s.Lease)
	require.Len(t, got, 1)
	require.Equal(t, entries[0].ID, got[0].ID)
	require.Equal(t, 2, got[0].Attempts)
	settle(t, repo, completed(got...)...)

	// Entries that are not processing are not extended.
	require.NoError(t, extender.Extend(t.Context(), s.Lease, entries[0].ID))
	requireStats(t, repo, gacha.Stats{Completed: 2})
}

func (s Suite[R]) testRedrive(t *testing.T) {
	repo := s.New(t)

	redriver, ok := repo.(gacha.Redriver)
	if !ok {
		t.Skip("gacha.Redriver not implemented")
	}

	values := s.values(0, 2)
	write(t, repo, values...)

	entries := claim(t, repo, 10, s.Lease)
	settle(t, repo, failed(entries[0], "boom"), completed(entries[1])[0])

	require.NoError(t, redriver.Redrive(t.Context()))
	// Duplicates, outbox entries and unknown IDs are ignored.
	require.NoError(t, redriver.Redrive(t.Context(), entries[0].ID, entries[0].ID, entries[1].ID, uuid.NewV7()))

	got := claim(t, repo, 10, s.Lease)
	require.Len(t, got, 1)
	require.Equal(t, entries[0].ID, got[0].ID)
	require.Equal(t, values[0], got[0].Data)
	require.Equal(t, 1, got[0].Attempts, "redriven entries start over")
	require.True(t, entries[0].CreatedAt.Equal(got[0].CreatedAt), "redriven entries keep their creation time")
	requireStats(t, repo, gacha.Stats{Processing: 1, Completed: 1})

	settle(t, repo, completed(got...)...)
	requireStats(t, repo, gacha.Stats{Completed: 2})
}

func (s Suite[R]) testStats(t *testing.T) {
	repo := s.New(t)

	if _, ok := repo.(gacha.Stater); !ok {
		t.Skip("gacha.Stater not implemented")
	}

	requireStats(t, repo, gacha.Stats{})

	write(t, repo, s.values(0, 4)...)
	requireStats(t, repo, gacha.Stats{Pending: 4})

	entries := claim(t, repo, 2, s.Lease)
	requireStats(t, repo, gacha.Stats{Pending: 2, Processing: 2})

	settle(t, repo, completed(entries[0])[0], failed(entries[1], "boom"))
	requireStats(t, repo, gacha.Stats{Pending: 2, Completed: 1, DeadLetters: 1})
}

func (s Suite[R]) testCompletedView(t *testing.T) {
	repo := s.New(t)

	viewer, ok := repo.(gacha.Viewer[R])
	if !ok {
		t.Skip("gacha.Viewer not implemented")
	}

	write(t, repo, s.values(0, 3)...)
	entries := claim(t, repo, 10, s.Lease)
	settle(t, repo, completed(entries...)...)

	// Pending entries are not part of the view.
	write(t, repo, s.Value(3))

	require.Empty(t, claim(t, viewer.Completed(time.Hour), 10, s.Lease), "retention hides recent entries")

	view := viewer.Completed(0)

	got := claim(t, view, 2, s.Lease)
	require.Equal(t, ids(entries[:2]), ids(got), "oldest first")
	require.Equal(t, data(entries[:2]), data(got))
	require.Equal(t, 1, got[0].Attempts, "views report the final attempt count")

	rest := claim(t, view, 10, s.Lease)
	require.Equal(t, ids(entries[2:]), ids(rest))
	require.Empty(t, claim(t, view, 10, s.Lease), "leased view entries are not claimed again")

	settle(t, view, completed(got...)...)
	requireStats(t, repo, gacha.Stats{Pending: 1, Completed: 1})

	s.expire()

	again := claim(t, view, 10, s.Lease)
	require.Equal(t, ids(rest), ids(again), "view leases expire")
}

func (s Suite[R]) testCompletedViewSettle(t *testing.T) {
	repo := s.New(t)

	viewer, ok := repo.(gacha.Viewer[R])
	if !ok {
		t.Skip("gacha.Viewer not implemented")
	}

	write(t, repo, s.values(0, 2)...)
	entries := claim(t, repo, 10, s.Lease)
	settle(t, repo, completed(entries[0])...)

	view := viewer.Completed(0)

	_, err := view.Claim(t.Context(), 1, 0)
	require.ErrorIs(t, err, gacha.ErrInvalidLease)
	require.ErrorIs(t, view.Settle(t.Context(), gacha.Outcome{ID: entries[0].ID}), gacha.ErrInvalidStatus)

	// A view only settles its own entries: the processing entry is untouched.
	settle(t, view, completed(entries[1])...)
	requireStats(t, repo, gacha.Stats{Processing: 1, Completed: 1})

	got := claim(t, view, 10, s.Lease)
	require.Equal(t, ids(entries[:1]), ids(got))

	// A retry makes the entry claimable again after RetryIn, and a release right away.
	settle(t, view, retry(got[0], 0, "archive failed"))

	got = claim(t, view, 10, s.Lease)
	require.Equal(t, ids(entries[:1]), ids(got))

	settle(t, view, released(got[0], ""))

	got = claim(t, view, 10, s.Lease)
	require.Equal(t, ids(entries[:1]), ids(got))
	require.Equal(t, 1, got[0].Attempts, "views do not count attempts")

	// A failure leaves the entry leased until the lease expires.
	settle(t, view, failed(got[0], "archive failed"))
	require.Empty(t, claim(t, view, 10, s.Lease))

	s.expire()

	require.Equal(t, ids(entries[:1]), ids(claim(t, view, 10, s.Lease)))

	settle(t, repo, completed(entries[1])...)
	requireStats(t, repo, gacha.Stats{Completed: 2})
}

func (s Suite[R]) testDeadLettersView(t *testing.T) {
	repo := s.New(t)

	viewer, ok := repo.(gacha.Viewer[R])
	if !ok {
		t.Skip("gacha.Viewer not implemented")
	}

	values := s.values(0, 2)
	write(t, repo, values...)

	entries := claim(t, repo, 10, s.Lease)
	settle(t, repo, failed(entries[0], "boom"), failed(entries[1], "boom"))

	require.Empty(t, claim(t, viewer.DeadLetters(time.Hour), 10, s.Lease), "retention hides recent entries")
	require.Empty(t, claim(t, viewer.Completed(0), 10, s.Lease), "dead letters are not completed")

	view := viewer.DeadLetters(0)

	got := claim(t, view, 10, s.Lease)
	require.Equal(t, ids(entries), ids(got))
	require.Equal(t, values, data(got))
	require.Equal(t, 1, got[0].Attempts)
	require.Empty(t, claim(t, view, 10, s.Lease))

	settle(t, view, completed(got[0])...)
	requireStats(t, repo, gacha.Stats{DeadLetters: 1})

	redriver, ok := repo.(gacha.Redriver)
	if !ok {
		return
	}

	// Purged dead letters are gone for good.
	require.NoError(t, redriver.Redrive(t.Context(), entries[0].ID, entries[1].ID))
	require.Equal(t, ids(entries[1:]), ids(claim(t, repo, 10, s.Lease)))
}

func (s Suite[R]) testDeleteOnComplete(t *testing.T) {
	if s.NewDeleteOnComplete == nil {
		t.Skip("Suite.NewDeleteOnComplete not set")
	}

	repo := s.NewDeleteOnComplete(t)
	write(t, repo, s.values(0, 3)...)

	entries := claim(t, repo, 10, s.Lease)
	settle(t, repo, completed(entries[0])[0], retry(entries[1], 0, ""), failed(entries[2], "boom"))
	requireStats(t, repo, gacha.Stats{Pending: 1, DeadLetters: 1})

	if viewer, ok := repo.(gacha.Viewer[R]); ok {
		require.Empty(t, claim(t, viewer.Completed(0), 10, s.Lease))
	}

	got := claim(t, repo, 10, s.Lease)
	require.Equal(t, ids(entries[1:2]), ids(got))
	settle(t, repo, completed(got...)...)

	// Settling a deleted entry is ignored.
	settle(t, repo, completed(entries[0])...)
	requireStats(t, repo, gacha.Stats{DeadLetters: 1})
}

func (s Suite[R]) testTxCommit(t *testing.T) {
	if s.Tx == nil {
		t.Skip("Suite.Tx not set")
	}

	repo := s.New(t)
	values := s.values(0, 2)

	sink, end := s.Tx(t, repo)
	require.NoError(t, sink.Write(t.Context(), values...))
	require.NoError(t, end(true))

	require.Equal(t, values, data(claim(t, repo, 10, s.Lease)))
}

func (s Suite[R]) testTxRollback(t *testing.T) {
	if s.Tx == nil {
		t.Skip("Suite.Tx not set")
	}

	repo := s.New(t)

	sink, end := s.Tx(t, repo)
	require.NoError(t, sink.Write(t.Context(), s.values(0, 2)...))
	require.NoError(t, end(false))

	require.Empty(t, claim(t, repo, 10, s.Lease))

	write(t, repo, s.Value(2))
	require.Equal(t, []R{s.Value(2)}, data(claim(t, repo, 10, s.Lease)))
}

// expire waits until the leases taken now have expired.
func (s Suite[R]) expire() {
	time.Sleep(s.Lease + s.Lease/2)
}

func (s Suite[R]) values(from, n int) []R {
	values := make([]R, n)

	for i := range values {
		values[i] = s.Value(from + i)
	}

	return values
}

func write[R any](t *testing.T, sink gacha.Sink[R], values ...R) {
	t.Helper()

	require.NoError(t, sink.Write(t.Context(), values...))
}

func claim[R any](t *testing.T, source gacha.Source[R], limit int, lease time.Duration) []gacha.Entry[R] {
	t.Helper()

	entries, err := source.Claim(t.Context(), limit, lease)
	require.NoError(t, err)

	return entries
}

func settle[R any](t *testing.T, source gacha.Source[R], outcomes ...gacha.Outcome) {
	t.Helper()

	require.NoError(t, source.Settle(t.Context(), outcomes...))
}

func requireStats[R any](t *testing.T, repo gacha.Repository[R], wants gacha.Stats) {
	t.Helper()

	stater, ok := repo.(gacha.Stater)
	if !ok {
		return
	}

	stats, err := stater.Stats(t.Context())
	require.NoError(t, err)
	require.Equal(t, wants, stats)
}

func completed[R any](entries ...gacha.Entry[R]) []gacha.Outcome {
	outcomes := make([]gacha.Outcome, len(entries))

	for i := range entries {
		outcomes[i] = gacha.Outcome{ID: entries[i].ID, Attempts: entries[i].Attempts, Status: gacha.StatusCompleted}
	}

	return outcomes
}

func retry[R any](entry gacha.Entry[R], delay time.Duration, err string) gacha.Outcome {
	return gacha.Outcome{
		ID:       entry.ID,
		Attempts: entry.Attempts,
		Status:   gacha.StatusPending,
		Err:      err,
		RetryIn:  delay,
	}
}

func released[R any](entry gacha.Entry[R], err string) gacha.Outcome {
	return gacha.Outcome{ID: entry.ID, Attempts: entry.Attempts, Status: gacha.StatusReleased, Err: err}
}

func failed[R any](entry gacha.Entry[R], err string) gacha.Outcome {
	return gacha.Outcome{ID: entry.ID, Attempts: entry.Attempts, Status: gacha.StatusFailed, Err: err}
}

func ids[R any](entries []gacha.Entry[R]) []uuid.UUID {
	out := make([]uuid.UUID, len(entries))

	for i := range entries {
		out[i] = entries[i].ID
	}

	return out
}

func data[R any](entries []gacha.Entry[R]) []R {
	out := make([]R, len(entries))

	for i := range entries {
		out[i] = entries[i].Data
	}

	return out
}
