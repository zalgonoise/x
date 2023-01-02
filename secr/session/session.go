package session

import "github.com/zalgonoise/x/secr/user"

type Session struct {
	User  user.User
	Token string
}
