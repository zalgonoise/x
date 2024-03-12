package config

import (
	"flag"
)

func Parse(args []string) (*Config, error) {
	fs := flag.NewFlagSet("authz", flag.ExitOnError)

	dbURI := fs.String("db", "", "the path to the SQLite DB file to store services and their certificates")
	privateKey := fs.String("private-key", "", "the path to the ECDSA private key file to use for the certificate authority")
	httpPort := fs.Int("http-port", 0, "the exposed HTTP port for the CA's API")
	grpcPort := fs.Int("grpc-port", 0, "the exposed gRPC port for the CA's API")
	dur := fs.Int("cert-dur", 0, "duration to use on new certificate's expiry")
	serviceName := fs.String("service-name", "", "service name to assign to a CA or Authz service")
	caURL := fs.String("ca-url", "", "address for the certificate authority that the Authz service should target")
	randomSize := fs.Int("rand-size", 0, "size for random numbers when generated for a challenge")
	challengeDur := fs.Duration("challenge-dur", 0, "duration for emitted challenges before they expire")
	tokenDur := fs.Duration("token-dur", 0, "duration for emitted tokens before they expire")
	cleanupTimeout := fs.Duration("cleanup-timeout", 0, "timeout duration when running DB cleanup on expired certificates")
	cleanupSchedule := fs.String("cleanup-cron", "", "cron schedule to run DB cleanup on expired certificates")
	tracerURL := fs.String("tracer-url", "", "URL for the tracing backend")
	tracerUsername := fs.String("tracer-username", "", "username for the tracing backend, if required")
	tracerPassword := fs.String("tracer-password", "", "password for the tracing backend, if required")
	tracerTimeout := fs.Duration("tracer-timeout", 0, "timeout when connecting to the tracing backend")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return &Config{
		PrivateKey: *privateKey,
		HTTPPort:   *httpPort,
		GRPCPort:   *grpcPort,
		Name:       *serviceName,
		CA: CA{
			CertDurMonths: *dur,
		},
		Authz: Authz{
			CAURL:         *caURL,
			RandSize:      *randomSize,
			CertDurMonths: *dur,
			ChallengeDur:  *challengeDur,
			TokenDur:      *tokenDur,
		},
		Database: Database{
			URI:             *dbURI,
			CleanupTimeout:  *cleanupTimeout,
			CleanupSchedule: *cleanupSchedule,
		},
		Tracer: Tracer{
			URI:         *tracerURL,
			Username:    *tracerUsername,
			Password:    *tracerPassword,
			ConnTimeout: *tracerTimeout,
		},
	}, nil
}
