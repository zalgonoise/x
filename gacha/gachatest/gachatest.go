// Package gachatest provides the conformance suite for gacha.Repository implementations.
//
// Every repository in this module runs it, and so should user implementations:
//
//	func TestRepository(t *testing.T) {
//		gachatest.TestRepository(t, func(t *testing.T) gacha.Repository[[]byte] {
//			return newRepository(t)
//		})
//	}
//
// Use Suite directly for a raw type other than []byte, or to enable the optional tests (transactional sinks,
// WithDeleteOnComplete) and tune timing. Optional capabilities (gacha.Extender, gacha.Redriver, gacha.Stater and
// gacha.Viewer) are tested when the repository implements them, and skipped otherwise.
package gachatest

import (
	"strconv"
	"testing"
	"time"

	"github.com/zalgonoise/x/gacha"
)

const defaultLease = 250 * time.Millisecond

// Suite is the conformance suite for gacha.Repository implementations over the raw type R.
type Suite[R any] struct {
	// New returns a new, empty repository. Required.
	New func(t *testing.T) gacha.Repository[R]

	// Value returns the i-th test value. Values for different i must differ. Required.
	Value func(i int) R

	// NewDeleteOnComplete, when set, returns a new, empty repository configured to delete entries on completion,
	// enabling the WithDeleteOnComplete tests.
	NewDeleteOnComplete func(t *testing.T) gacha.Repository[R]

	// Tx, when set, enables the transactional sink tests. It begins a transaction against repo and returns a sink
	// bound to it, along with a function that commits (true) or rolls back (false) the transaction.
	Tx func(t *testing.T, repo gacha.Repository[R]) (sink gacha.Sink[R], end func(commit bool) error)

	// Lease is the lease of the time-dependent tests, which wait up to a few leases. It defaults to 250ms: raise it
	// for backends with a coarse clock, and keep it when Wrap fakes the clock.
	Lease time.Duration

	// Wrap, when set, wraps each test. Set it to synctest.Test for repositories that use the process clock, which
	// makes the time-dependent tests instant and deterministic.
	Wrap func(t *testing.T, test func(t *testing.T))
}

// TestRepository runs the conformance suite against the repositories returned by newRepo, which must be new and
// empty on every call.
func TestRepository(t *testing.T, newRepo func(t *testing.T) gacha.Repository[[]byte]) {
	t.Helper()

	Suite[[]byte]{New: newRepo, Value: Bytes}.Run(t)
}

// Bytes returns distinct []byte test values, for Suite.Value.
func Bytes(i int) []byte {
	return []byte("value-" + strconv.Itoa(i))
}

// Run runs the suite as subtests of t.
func (s Suite[R]) Run(t *testing.T) {
	t.Helper()

	if s.New == nil || s.Value == nil {
		t.Fatal("gachatest: Suite.New and Suite.Value are required")
	}

	if s.Lease <= 0 {
		s.Lease = defaultLease
	}

	for _, test := range []struct {
		name string
		run  func(Suite[R], *testing.T)
	}{
		{name: "Write/Empty", run: Suite[R].testWriteEmpty},
		{name: "Write/CanceledContext", run: Suite[R].testCanceledContext},
		{name: "Claim/RoundTrip", run: Suite[R].testRoundTrip},
		{name: "Claim/Order", run: Suite[R].testClaimOrder},
		{name: "Claim/Limits", run: Suite[R].testClaimLimits},
		{name: "Claim/InvalidLease", run: Suite[R].testClaimInvalidLease},
		{name: "Claim/LeaseExpiry", run: Suite[R].testLeaseExpiry},
		{name: "Claim/Concurrent", run: Suite[R].testConcurrentClaims},
		{name: "Settle/Empty", run: Suite[R].testSettleEmpty},
		{name: "Settle/Completed", run: Suite[R].testSettleCompleted},
		{name: "Settle/Retry", run: Suite[R].testSettleRetry},
		{name: "Settle/RetryImmediately", run: Suite[R].testSettleRetryImmediately},
		{name: "Settle/Failed", run: Suite[R].testSettleFailed},
		{name: "Settle/Mixed", run: Suite[R].testSettleMixed},
		{name: "Settle/Fenced", run: Suite[R].testSettleFenced},
		{name: "Settle/Released", run: Suite[R].testSettleReleased},
		{name: "Settle/Ignored", run: Suite[R].testSettleIgnored},
		{name: "Settle/InvalidStatus", run: Suite[R].testSettleInvalidStatus},
		{name: "Extend", run: Suite[R].testExtend},
		{name: "Redrive", run: Suite[R].testRedrive},
		{name: "Stats", run: Suite[R].testStats},
		{name: "Views/Completed", run: Suite[R].testCompletedView},
		{name: "Views/CompletedSettle", run: Suite[R].testCompletedViewSettle},
		{name: "Views/DeadLetters", run: Suite[R].testDeadLettersView},
		{name: "DeleteOnComplete", run: Suite[R].testDeleteOnComplete},
		{name: "Tx/Commit", run: Suite[R].testTxCommit},
		{name: "Tx/Rollback", run: Suite[R].testTxRollback},
	} {
		t.Run(test.name, func(t *testing.T) {
			if s.Wrap == nil {
				test.run(s, t)

				return
			}

			s.Wrap(t, func(t *testing.T) {
				test.run(s, t)
			})
		})
	}
}
