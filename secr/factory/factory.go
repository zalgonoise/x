package factory

import (
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/x/secr/cmd/config"
	"github.com/zalgonoise/x/secr/service"
	"github.com/zalgonoise/x/secr/transport/http"
)

// Service creates a new service based on the signing key path `authKeyPath`,
// Bolt DB path `boltDBPath`, and SQLite DB path `sqliteDBPath`
func Service(authKeyPath, boltDBPath, sqliteDBPath string) (service.Service, error) {
	authorizer, err := Authorizer(authKeyPath)
	if err != nil {
		return nil, err
	}

	keys, err := Bolt(boltDBPath)
	if err != nil {
		return nil, err
	}

	users, secrets, err := SQLite(sqliteDBPath)
	if err != nil {
		return nil, err
	}

	return service.WithLogger(logx.Default(), service.WithTrace(service.NewService(
		users, secrets, keys, authorizer,
	))), nil
}

// Server creates a new HTTP server based on the service created using the
// signing key path `authKeyPath`, Bolt DB path `boltDBPath`, and SQLite DB path `sqliteDBPath`
func Server(port int, authKeyPath, boltDBPath, sqliteDBPath string) (http.Server, error) {
	if port == 0 {
		port = config.Default.HTTPPort
	}

	authorizer, err := Authorizer(authKeyPath)
	if err != nil {
		return nil, err
	}

	keys, err := Bolt(boltDBPath)
	if err != nil {
		return nil, err
	}

	users, secrets, err := SQLite(sqliteDBPath)
	if err != nil {
		return nil, err
	}

	return http.NewServer(
		port,
		service.WithLogger(logx.Default(),
			service.WithTrace(service.NewService(
				users, secrets, keys, authorizer,
			))),
	), nil
}

// From creates a HTTP server for the Secrets service based on the input config
func From(conf *config.Config) (http.Server, error) {
	authorizer, err := Authorizer(conf.SigningKeyPath)
	if err != nil {
		return nil, err
	}

	keys, err := Bolt(conf.BoltDBPath)
	if err != nil {
		return nil, err
	}

	users, secrets, err := SQLite(conf.SQLiteDBPath)
	if err != nil {
		return nil, err
	}

	svc := service.WithLogger(logx.Default(), service.WithTrace(service.NewService(
		users, secrets, keys, authorizer,
	)))

	return http.NewServer(conf.HTTPPort, svc), nil
}
