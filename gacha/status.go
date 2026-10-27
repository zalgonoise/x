package gacha

import "strconv"

// Status is the state of an outbox entry, and the verdict an Outcome carries for it.
//
// The zero value is not a valid status, so a zero Outcome is rejected by Source.Settle instead of silently meaning
// "retry".
type Status uint8

const (
	// StatusPending marks an entry waiting to be claimed. As an Outcome, it requests a retry after Outcome.RetryIn.
	StatusPending Status = iota + 1
	// StatusProcessing marks a claimed entry whose lease has not been settled yet. It is not a valid Outcome.
	StatusProcessing
	// StatusCompleted marks an entry whose handler succeeded.
	StatusCompleted
	// StatusFailed marks an entry moved to the dead-letter table. As an Outcome, it requests that move.
	StatusFailed
	// StatusReleased is only an Outcome, never an entry state: it hands a claimed entry back as pending, claimable
	// right away, without counting the claim as an attempt. The poller releases the entries a shutdown kept it from
	// handling.
	StatusReleased
)

const (
	pendingText    = "pending"
	processingText = "processing"
	completedText  = "completed"
	failedText     = "failed"
	releasedText   = "released"
)

// String returns the lowercase name of the status, as stored by the SQL repositories.
func (s Status) String() string {
	switch s {
	case StatusPending:
		return pendingText
	case StatusProcessing:
		return processingText
	case StatusCompleted:
		return completedText
	case StatusFailed:
		return failedText
	case StatusReleased:
		return releasedText
	default:
		return "Status(" + strconv.Itoa(int(s)) + ")"
	}
}

// IsValid reports whether s is one of the defined statuses.
func (s Status) IsValid() bool {
	return s >= StatusPending && s <= StatusReleased
}

// ParseStatus returns the Status named by text, as returned by Status.String. It returns ErrInvalidStatus for any
// other input.
func ParseStatus(text string) (Status, error) {
	switch text {
	case pendingText:
		return StatusPending, nil
	case processingText:
		return StatusProcessing, nil
	case completedText:
		return StatusCompleted, nil
	case failedText:
		return StatusFailed, nil
	case releasedText:
		return StatusReleased, nil
	default:
		return 0, ErrInvalidStatus
	}
}

// MarshalText implements encoding.TextMarshaler. It returns ErrInvalidStatus for an invalid status.
func (s Status) MarshalText() ([]byte, error) {
	if !s.IsValid() {
		return nil, ErrInvalidStatus
	}

	return []byte(s.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler, accepting the forms returned by Status.String.
func (s *Status) UnmarshalText(text []byte) error {
	status, err := ParseStatus(string(text))
	if err != nil {
		return err
	}

	*s = status

	return nil
}
