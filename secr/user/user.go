package user

import "time"

type User struct {
	ID        uint64
	Username  string
	Name      string
	Hash      string
	Salt      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
