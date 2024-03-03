package authz

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/cfg"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	csr             *pb.CSR
	challengeExpiry time.Duration
	tokenExpiry     time.Duration

	m      Metrics
	logger *slog.Logger
	tracer trace.Tracer
}

func WithChallengeExpiry(dur time.Duration) cfg.Option[Config] {
	if dur <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.challengeExpiry = dur

		return config
	})
}

func WithTokenExpiry(dur time.Duration) cfg.Option[Config] {
	if dur <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.tokenExpiry = dur

		return config
	})
}

func WithCSR(csr *pb.CSR) cfg.Option[Config] {
	if csr == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.csr = csr

		return config
	})
}

func WithMetrics(m Metrics) cfg.Option[Config] {
	if m == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.m = m

		return config
	})
}

func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	if logger == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.logger = logger

		return config
	})
}

func WithTracer(tracer trace.Tracer) cfg.Option[Config] {
	if tracer == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.tracer = tracer

		return config
	})
}
