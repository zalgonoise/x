package shared

import (
	"errors"
	"time"
)

var zeroTime = time.Time{}

var (
	ErrEmptyDuration = errors.New("duration cannot be zero")
	ErrEmptyTime     = errors.New("time cannot be zero")
	ErrExpired       = errors.New("input time is already expired")
)

func ValidateDuration(dur time.Duration) error {
	if dur == 0 {
		return ErrEmptyDuration
	}
	return nil
}

func ValidateTime(t time.Time) error {
	if t.IsZero() || t == zeroTime {
		return ErrEmptyTime
	}
	if time.Now().After(t) {
		return ErrExpired
	}
	return nil
}
