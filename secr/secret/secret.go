package secret

import "github.com/zalgonoise/x/secr/user"

type Secret struct {
	Key   string
	Value []byte
	Owner user.User
}
