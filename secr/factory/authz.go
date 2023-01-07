package factory

import (
	"errors"
	"os"

	"github.com/zalgonoise/x/secr/authz"
	"github.com/zalgonoise/x/secr/crypt"
)

const (
	signingKeyPath = "/secr/server/key"
)

func Authorizer(key []byte) (authz.Authorizer, error) {
	if len(key) > 0 {
		return authz.NewAuthorizer(key), nil
	}

	fs, err := os.Stat(signingKeyPath)
	if (err != nil && os.IsNotExist(err)) || (fs != nil && fs.Size() == 0) {
		// try to create local key
		k := crypt.NewKey()
		f, err := os.Create(signingKeyPath)
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

	// try to load local key
	f, err := os.Open(signingKeyPath)
	if err != nil {
		return nil, err
	}
	var k = make([]byte, 0, 256)
	n, err := f.Read(k)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, errors.New("zero bytes written")
	}

	return authz.NewAuthorizer(k), nil

}
