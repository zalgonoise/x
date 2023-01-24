package user

import (
	"errors"
	"regexp"
)

var (
	ErrEmptyUsername   = errors.New("username cannot be empty")
	ErrShortUsername   = errors.New("username is too short")
	ErrLongUsername    = errors.New("username is too long")
	ErrInvalidUsername = errors.New("invalid username")

	ErrEmptyName   = errors.New("name cannot be empty")
	ErrShortName   = errors.New("name is too short")
	ErrLongName    = errors.New("name is too long")
	ErrInvalidName = errors.New("invalid name")

	ErrEmptyPassword   = errors.New("password cannot be empty")
	ErrShortPassword   = errors.New("password is too short")
	ErrLongPassword    = errors.New("password is too long")
	ErrInvalidPassword = errors.New("invalid password")
)

const (
	usernameMinLength = 3
	usernameMaxLength = 25

	nameMinLength = 2
	nameMaxLength = 25

	passwordMinLength       = 7
	passwordMaxLength       = 300
	PasswordCharRepeatLimit = 4
	passwordAllowedChars    = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_~!@#$%^&*()=[]{}'\"|,./<>?;:`
)

var (
	usernameRegex = regexp.MustCompile(`[a-z0-9]+[a-z0-9\-_]+[a-z0-9]+`)
	nameRegex     = regexp.MustCompile(`[a-zA-Z]+[\s]?[a-zA-Z]+`)

	passwordCharMap = map[rune]struct{}{}
)

func init() {
	for _, c := range passwordAllowedChars {
		passwordCharMap[c] = struct{}{}
	}
}

// ValidateUsername verifies if the input username is valid, returning an error
// if invalid
func ValidateUsername(username string) error {
	if username == "" {
		return ErrEmptyUsername
	}
	if len(username) < usernameMinLength {
		return ErrShortUsername
	}
	if len(username) > usernameMaxLength {
		return ErrLongUsername
	}
	if match := usernameRegex.FindString(username); match != username || username == RootUsername {
		return ErrInvalidUsername
	}
	return nil
}

// ValidateName verifies if the input name is valid, returning an error
// if invalid
func ValidateName(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	if len(name) < nameMinLength {
		return ErrShortName
	}
	if len(name) > nameMaxLength {
		return ErrLongName
	}
	if match := nameRegex.FindString(name); match != name {
		return ErrInvalidName
	}
	return nil
}

// ValidatePassword verifies if the input password is valid, returning an error
// if invalid
func ValidatePassword(password string) error {
	if password == "" {
		return ErrEmptyPassword
	}
	if len(password) < passwordMinLength {
		return ErrShortPassword
	}
	if len(password) > passwordMaxLength {
		return ErrLongPassword
	}
	return validatePasswordCharacters(password)
}

func validatePasswordCharacters(password string) error {
	var cur rune
	var count int = 1
	for _, c := range password {
		switch c {
		case cur:
			count++
			if count >= PasswordCharRepeatLimit {
				return ErrInvalidPassword
			}
		default:
			cur = c
			count = 1
		}
		if _, ok := passwordCharMap[c]; !ok {
			return ErrInvalidPassword
		}
	}
	return nil
}
