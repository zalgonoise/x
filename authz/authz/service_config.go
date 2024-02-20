package authz

import (
	"log/slog"

	"github.com/zalgonoise/cfg"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	csr *pb.CSR

	m      Metrics
	logger *slog.Logger
	tracer trace.Tracer
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
