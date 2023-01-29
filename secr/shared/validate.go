package shared

import (
	"time"

	"github.com/zalgonoise/x/errors"
)

var zeroTime = time.Time{}

var (
	ErrEmptyDuration = errors.New("duration cannot be zero")
	ErrEmptyTime     = errors.New("time cannot be zero")
	ErrExpired       = errors.New("input time is already expired")
)

// ValidateDuration verifies if the input duration is valid, returning an error
// if otherwise
func ValidateDuration(dur time.Duration) error {
	if dur == 0 {
		return ErrEmptyDuration
	}
	return nil
}

// ValidateTime verifies if the input time is valid, returning an error
// if otherwise
func ValidateTime(t time.Time) error {
	if t.IsZero() || t == zeroTime {
		return ErrEmptyTime
	}
	if time.Now().After(t) {
		return ErrExpired
	}
	return nil
}
