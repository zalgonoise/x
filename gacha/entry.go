package gacha

import (
	"time"
	"uuid"
)

// Entry is an outbox entry as returned by Source.Claim.
type Entry[R any] struct {
	// ID identifies the entry. Repositories generate it as a UUIDv7 on write, and it is kept unchanged when the entry
	// moves to the dead-letter table and back.
	ID uuid.UUID
	// Data is the raw, stored form of the item.
	Data R
	// Attempts is the number of times the entry has been claimed, including the current claim. It is the fencing
	// token to set in the Outcome that settles this claim.
	Attempts int
	// CreatedAt is the time the entry was first written.
	CreatedAt time.Time
}

// Outcome is the verdict for a claimed entry, applied by Source.Settle.
type Outcome struct {
	// ID identifies the entry to settle.
	ID uuid.UUID
	// Attempts is the Entry.Attempts value of the claim being settled. Source.Settle ignores the outcome when the
	// entry has been claimed again since (its lease expired and another worker took it over), so a stale worker never
	// overwrites the outcome of the current one.
	Attempts int
	// Status is the verdict: StatusCompleted, StatusPending (retry after RetryIn), StatusFailed (move the entry to
	// the dead-letter table) or StatusReleased (hand the entry back without counting the claim as an attempt).
	Status Status
	// Err is recorded as the entry's last error when not empty.
	Err string
	// RetryIn is the delay before a retried entry (StatusPending) can be claimed again. Negative values are treated as
	// zero.
	RetryIn time.Duration
}

// Validate returns ErrInvalidStatus when the outcome's status is not StatusCompleted, StatusPending, StatusFailed or
// StatusReleased. Sources call it on every outcome before applying any.
func (o Outcome) Validate() error {
	if !o.Status.IsValid() || o.Status == StatusProcessing {
		return ErrInvalidStatus
	}

	return nil
}

// Stats holds the number of entries in each state, as returned by Stater.Stats.
type Stats struct {
	// Pending is the number of entries waiting to be claimed, including retries waiting for their backoff.
	Pending int64
	// Processing is the number of claimed entries that have not been settled, including those with expired leases.
	Processing int64
	// Completed is the number of completed entries not purged yet.
	Completed int64
	// DeadLetters is the number of entries in the dead-letter table.
	DeadLetters int64
}
