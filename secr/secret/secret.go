package secret

import (
	"time"

	"github.com/zalgonoise/x/secr/user"
)

type Secret struct {
	Key       string
	Value     []byte
	CreatedAt time.Time
}

type WithOwner struct {
	Secret
	user.User
}

type Shared struct {
	Secret
	Owner  user.User
	Shares []user.User
	Until  time.Time
}
