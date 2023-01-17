package user

import (
	"time"
)

// RootUsername is a reserved username for the admin deploying the application
const RootUsername = "root"

// User is a person (or entity) that uses the application to store
// secrets. They will have a unique username.
type User struct {
	ID        uint64
	Username  string
	Name      string
	Hash      string
	Salt      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Session is an authorized user, accompanied by a JWT
type Session struct {
	User
	Token string
}
