package factory

import "github.com/zalgonoise/x/secr/service"

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
