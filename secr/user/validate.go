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

	passwordMinLength = 7
	passwordMaxLength = 30
)

var (
	usernameRegex = regexp.MustCompile(`[a-z0-9]+[a-z0-9\-_]+[a-z0-9]+`)
	nameRegex     = regexp.MustCompile(`[a-zA-Z]+[\s]?[a-zA-Z]+`)
	passwordRegex = regexp.MustCompile(`[a-zA-Z\d\-\_\~\!\@\#\$\%\^\&\*\(\)\[\]\{\}\;\:\,\.\<\>\?]+`)
)

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
	if match := passwordRegex.FindString(password); match != password {
		return ErrInvalidPassword
	}
	return nil
}
