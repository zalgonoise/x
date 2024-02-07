package ca

import (
	"crypto/x509"
	"log/slog"

	"github.com/zalgonoise/cfg"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	logHandler slog.Handler
	tracer     trace.Tracer

	cert *x509.Certificate
}

func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	if logger == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.logHandler = logger.Handler()

		return config
	})
}

func WithLogHandler(logger slog.Handler) cfg.Option[Config] {
	if logger == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.logHandler = logger

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

func WithTemplate(certificate *x509.Certificate) cfg.Option[Config] {
	if certificate == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.cert = certificate

		return config
	})
}
