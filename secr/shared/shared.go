package shared

import (
	"time"

	"github.com/zalgonoise/x/secr/user"
)

// Shared is metadata for a secret that a user (the owner) shares with a set of users
// optionally within a limited period of time
type Shared struct {
	Key    string
	Owner  user.User
	Shares []user.User
	Until  time.Time
}
