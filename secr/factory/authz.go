package factory

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/zalgonoise/x/secr/authz"
	"github.com/zalgonoise/x/secr/crypt"
)

const (
	signingKeyPath = "/secr/server/key"
)

// Authorizer creates a new authorizer with the input key `key`, or creates a new
// one under the preset folder if it doesn't yet exist and is not provided
func Authorizer(path string) (auth authz.Authorizer, err error) {
	var defErr error

	auth, err = loadKey(path)
	if err != nil {
		auth, defErr = loadKey(signingKeyPath)
		if defErr != nil {
			return nil, fmt.Errorf("failed to load key from input path and from default path: %w ; %v", err, defErr)
		}
	}
	return authz.WithTrace(auth), nil
}

func loadKey(path string) (authz.Authorizer, error) {
	fs, err := os.Stat(path)
	if (err != nil && os.IsNotExist(err)) || (fs != nil && fs.Size() == 0) {
		return createKey(path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(f)

	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, errors.New("zero bytes read")
	}

	return authz.NewAuthorizer(b), nil
}

func createKey(path string) (authz.Authorizer, error) {
	// try to create local key
	k := crypt.New256Key()
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	n, err := f.Write(k[:])
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, errors.New("zero bytes written")
	}
	return authz.NewAuthorizer(k[:]), nil
}
