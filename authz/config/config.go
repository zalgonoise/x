package config

import (
	"time"
)

const (
	defaultHTTPPort          = 8080
	defaultGRPCPort          = 8081
	defaultCACertDur         = 12
	defaultServiceCertDur    = 12
	defaultAuthzName         = "authz"
	defaultRandomSize        = 32
	defaultChallengeDur      = 10 * time.Minute
	defaultTokenDur          = 2 * time.Hour
	defaultDBCleanupTimeout  = 5 * time.Minute
	defaultDBCleanupSchedule = "0 6 * * *"
	defaultTracerTimeout     = 2 * time.Minute
)

type Config struct {
	PrivateKey string `envconfig:"AUTHZ_PRIVATE_KEY_PATH"`
	HTTPPort   int    `envconfig:"AUTHZ_HTTP_PORT"`
	GRPCPort   int    `envconfig:"AUTHZ_GRPC_PORT"`
	CA         CA
	Authz      Authz
	Database   Database
	Tracer     Tracer
}

type CA struct {
	CertDurMonths int `envconfig:"AUTHZ_CA_CERT_DUR_MOTNHS"`
}

type Authz struct {
	Name          string        `envconfig:"AUTHZ_SERVICE_NAME"`
	CAURL         string        `envconfig:"AUTHZ_TARGET_CA_URL"`
	RandSize      int           `envconfig:"AUTHZ_RANDOM_INT_SIZE"`
	CertDurMonths int           `envconfig:"AUTHZ_SERVICE_CERT_DUR_MOTNHS"`
	ChallengeDur  time.Duration `envconfig:"AUTHZ_CHALLENGE_DURATION"`
	TokenDur      time.Duration `envconfig:"AUTHZ_TOKEN_DURATION"`
}

type Database struct {
	URI             string        `envconfig:"AUTHZ_DATABASE_URI"`
	CleanupTimeout  time.Duration `envconfig:"AUTHZ_DATABASE_CLEANUP_TIMEOUT"`
	CleanupSchedule string        `envconfig:"AUTHZ_DATABASE_CLEANUP_SCHEDULE"`
}

type Tracer struct {
	URI         string        `envconfig:"AUTHZ_TRACER_URI"`
	Username    string        `envconfig:"AUTHZ_TRACER_USERNAME"`
	Password    string        `envconfig:"AUTHZ_TRACER_PASSWORD"`
	ConnTimeout time.Duration `envconfig:"AUTHZ_TRACER_CONNECTION_TIMEOUT"`
}

func defaultConfig() *Config {
	return &Config{
		HTTPPort: defaultHTTPPort,
		GRPCPort: defaultGRPCPort,
		CA: CA{
			CertDurMonths: defaultCACertDur,
		},
		Authz: Authz{
			Name:          defaultAuthzName,
			RandSize:      defaultRandomSize,
			CertDurMonths: defaultServiceCertDur,
			ChallengeDur:  defaultChallengeDur,
			TokenDur:      defaultTokenDur,
		},
		Database: Database{
			CleanupTimeout:  defaultDBCleanupTimeout,
			CleanupSchedule: defaultDBCleanupSchedule,
		},
		Tracer: Tracer{
			ConnTimeout: defaultTracerTimeout,
		},
	}
}

func New(args []string) (*Config, error) {
	config := defaultConfig()

	env, err := Get()
	if err != nil {
		return nil, err
	}

	flags, err := Parse(args)
	if err != nil {
		return nil, err
	}

	config = Merge(config, env)

	return Merge(config, flags), nil
}

func Merge(cur, next *Config) *Config {
	// top-level
	if next.PrivateKey != "" {
		cur.PrivateKey = next.PrivateKey
	}

	if next.HTTPPort > 0 {
		cur.HTTPPort = next.HTTPPort
	}

	if next.GRPCPort > 0 {
		cur.GRPCPort = next.GRPCPort
	}

	// CA
	if next.CA.CertDurMonths > 0 {
		cur.CA.CertDurMonths = next.CA.CertDurMonths
	}

	// Authz

	// Database
	if next.Database.URI != "" {
		cur.Database.URI = next.Database.URI
	}

	if next.Database.CleanupTimeout > 0 {
		cur.Database.CleanupTimeout = next.Database.CleanupTimeout
	}

	if next.Database.CleanupSchedule != "" {
		cur.Database.CleanupSchedule = next.Database.CleanupSchedule
	}

	// Tracer
	if next.Tracer.URI != "" {
		cur.Tracer.URI = next.Tracer.URI
	}

	if next.Tracer.Username != "" {
		cur.Tracer.Username = next.Tracer.Username
	}

	if next.Tracer.Password != "" {
		cur.Tracer.Password = next.Tracer.Password
	}

	if next.Tracer.ConnTimeout > 0 {
		cur.Tracer.ConnTimeout = next.Tracer.ConnTimeout
	}

	return cur
}
