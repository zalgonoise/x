package secret

import (
	"time"
)

// Secret is a key-value pair where they Key is string type and Value
// is a slice of bytes. Secrets are encrypted then stored with a user-scoped
// private key
type Secret struct {
	ID        uint64    `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}
