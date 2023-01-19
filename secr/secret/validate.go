package secret

import (
	"errors"
	"regexp"
)

var (
	ErrEmptyKey   = errors.New("key cannot be empty")
	ErrLongKey    = errors.New("key is too long")
	ErrInvalidKey = errors.New("invalid key")

	ErrEmptyValue = errors.New("value cannot be empty")
	ErrLongValue  = errors.New("value is too long")
)

const (
	keyMaxLength   = 20
	valueMaxLength = 8192
)

var (
	keyRegex = regexp.MustCompile(`[a-z0-9]+[a-z0-9\-_]+[a-z0-9]+`)
)

// ValidateKey verifies if the input secret's key is valid, returning an error
// if invalid
func ValidateKey(key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	if len(key) > keyMaxLength {
		return ErrLongKey
	}
	if match := keyRegex.FindString(key); match != key {
		return ErrInvalidKey
	}
	return nil
}

// ValidateValue verifies if the input secret's value is valid, returning an error
// if invalid
func ValidateValue(value []byte) error {
	if len(value) == 0 {
		return ErrEmptyValue
	}
	if len(value) > valueMaxLength {
		return ErrLongValue
	}
	return nil
}
