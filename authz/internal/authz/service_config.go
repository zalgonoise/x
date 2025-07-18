package authz

import (
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/authz/internal/log"
	"github.com/zalgonoise/x/authz/internal/metrics"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
)

type Config struct {
	csr             *pb.CSR
	durMonth        int
	challengeExpiry time.Duration
	tokenExpiry     time.Duration

	m      Metrics
	logger slog.Handler
	tracer trace.Tracer
}

func defaultConfig() Config {
	return Config{
		durMonth:        defaultDurMonth,
		challengeExpiry: defaultChallengeExpiry,
		tokenExpiry:     defaultTokenExpiry,
		m:               metrics.NoOp(),
		logger:          log.NoOp().Handler(),
		tracer:          noop.NewTracerProvider().Tracer("authz"),
	}
}

func WithDurMonth(months int) cfg.Option[Config] {
	if months <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.durMonth = months

		return config
	})
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
		config.logger = logger.Handler()

		return config
	})
}

func WithLogHandler(handler slog.Handler) cfg.Option[Config] {
	if handler == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.logger = handler

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
