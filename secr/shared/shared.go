package shared

import (
	"time"
)

// Shared is metadata for a secret that a user (the owner) shares with a set of users
// optionally within a limited period of time
type Share struct {
	ID        uint64     `json:"id"`
	SecretKey string     `json:"secret_key"`
	Owner     string     `json:"owner"`
	Target    []string   `json:"targets"`
	Until     *time.Time `json:"until,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
