package secret

import (
	"time"

	"github.com/zalgonoise/x/secr/user"
)

// Secret is a key-value pair where they Key is string type and Value
// is a slice of bytes. Secrets are encrypted then stored with a user-scoped
// private key
type Secret struct {
	Key       string
	Value     []byte
	CreatedAt time.Time
}

// WithOwner is a type of secret that is tied to a certain user as an owner
type WithOwner struct {
	Secret
	user.User
}

// Shared is a type of secret that a user (the owner) shares with a set of users
// optionally within a limited period of time
type Shared struct {
	Secret
	Owner  user.User
	Shares []user.User
	Until  time.Time
}
