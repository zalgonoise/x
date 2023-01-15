package factory

import (
	"github.com/zalgonoise/x/secr/cmd/config"
	"github.com/zalgonoise/x/secr/service"
	"github.com/zalgonoise/x/secr/transport/http"
)

// Service creates a new service based on the signing key path `authKeyPath`,
// Bolt DB path `boltDBPath`, and SQLite DB path `sqliteDBPath`
func Service(authKeyPath, boltDBPath, sqliteDBPath string) (service.Service, error) {
	Spanner(traceFilePath)

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

	return service.NewService(
		users, secrets, keys, authorizer,
	), nil
}

// Service creates a new service based on the signing key path `authKeyPath`,
// Bolt DB path `boltDBPath`, and SQLite DB path `sqliteDBPath`
func WithLogAndTrace(traceFilePath, logFilePath string, svc service.Service) service.Service {
	Spanner(traceFilePath)
	return service.WithLogger(
		Logger(logFilePath),
		service.WithTrace(svc),
	)
}

// Server creates a new HTTP server based on the service created using the
// signing key path `authKeyPath`, Bolt DB path `boltDBPath`, and SQLite DB path `sqliteDBPath`
func Server(port int, svc service.Service) (http.Server, error) {
	if port == 0 {
		port = config.Default.HTTPPort
	}

	return http.NewServer(
		port,
		svc,
	), nil
}

// From creates a HTTP server for the Secrets service based on the input config
func From(conf *config.Config) (http.Server, error) {
	svc, err := Service(conf.SigningKeyPath, conf.BoltDBPath, conf.SQLiteDBPath)
	if err != nil {
		return nil, err
	}

	loggedSvc := WithLogAndTrace(
		conf.TraceFilePath,
		conf.LogFilePath,
		svc,
	)

	return Server(conf.HTTPPort, loggedSvc)
}
