package shared

import (
	"time"

	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/user"
)

// Shared is metadata for a secret that a user (the owner) shares with a set of users
// optionally within a limited period of time
type Share struct {
	ID        uint64
	Secret    secret.Secret
	Owner     user.User
	Target    []user.User
	Until     *time.Time
	CreatedAt time.Time
}
