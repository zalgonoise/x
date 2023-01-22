package shared

import (
	"time"

	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/user"
)

// Shared is metadata for a secret that a user (the owner) shares with a set of users
// optionally within a limited period of time
type Share struct {
	ID        uint64        `json:"id"`
	Secret    secret.Secret `json:"secret"`
	Owner     user.User     `json:"owner"`
	Target    []user.User   `json:"targets"`
	Until     *time.Time    `json:"until,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}
