package user

import (
	"time"
)

// User is a person (or entity) that uses the application to store
// secrets. They will have a unique username.
type User struct {
	ID        uint64
	Username  string
	Name      string
	Password  string
	Hash      string
	Salt      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPassword represents a password-changing user. The password within the user object
// must be the current, while the one within the NewPassword object should be the new one
type NewPassword struct {
	User     User
	Password string
}

// Session is an authorized user, accompanied by a JWT
type Session struct {
	User
	Token string
}
