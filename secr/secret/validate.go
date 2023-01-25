package secret

import (
	"errors"
	"regexp"
	"strings"

	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrEmptyKey   = errors.New("key cannot be empty")
	ErrLongKey    = errors.New("key is too long")
	ErrInvalidKey = errors.New("invalid key")

	ErrEmptyValue = errors.New("value cannot be empty")
	ErrLongValue  = errors.New("value is too long")

	ErrInvalidSharedKey = errors.New("invalid shared key")
)

const (
	keyMaxLength   = 20
	valueMaxLength = 8192
)

var (
	keyRegex = regexp.MustCompile(`[a-z0-9]+[a-z0-9\-_:]+[a-z0-9]+`)
)

// ValidateKey verifies if the input secret's key is valid, returning an error
// if invalid
func ValidateKey(key string) (bool, error) {
	if key == "" {
		return false, ErrEmptyKey
	}
	if len(key) > keyMaxLength {
		return false, ErrLongKey
	}
	if match := keyRegex.FindString(key); match != key {
		return false, ErrInvalidKey
	}
	if key == keys.UniqueID || key == keys.TokenKey {
		return false, ErrEmptyKey
	}
	if strings.Contains(key, ":") {
		split := strings.SplitN(key, ":", 1)
		if len(split) != 2 {
			return false, ErrInvalidSharedKey
		}
		if err := user.ValidateUsername(split[0]); err != nil {
			return false, err
		}
		if isShared, err := ValidateKey(split[1]); err != nil || isShared {
			return false, err
		}
		return true, nil
	}
	return false, nil
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
