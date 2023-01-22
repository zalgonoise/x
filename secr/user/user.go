package user

import (
	"time"
)

// RootUsername is a reserved username for the admin deploying the application
const RootUsername = "root"

// User is a person (or entity) that uses the application to store
// secrets. They will have a unique username.
type User struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Hash      string    `json:"-"`
	Salt      string    `json:"-"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Session is an authorized user, accompanied by a JWT
type Session struct {
	User  `json:"user"`
	Token string `json:"token"`
}
