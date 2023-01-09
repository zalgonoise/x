package factory

import "github.com/zalgonoise/x/secr/service"

// Service creates a new service based on the signing key `key`
func Service(key []byte) (service.Service, error) {
	authorizer, err := Authorizer(key)
	if err != nil {
		return nil, err
	}

	keys, err := Bolt()
	if err != nil {
		return nil, err
	}

	users, secrets, err := SQLite()
	if err != nil {
		return nil, err
	}

	return service.NewService(
		users, secrets, keys, authorizer,
	), nil
}
