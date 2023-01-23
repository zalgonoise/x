package keys

import "fmt"

const (
	UniqueID = "unique_identifier"
	TokenKey = "active-token"

	// reserved: will be used encrypting user private keys
	// TODO: implement this ^
	ServerID = "secr-server-id"
)

// UserBucket formats the input user ID as a user bucket identifier (`uid:###`)
func UserBucket(id uint64) string {
	return fmt.Sprintf("uid:%d", id)
}
