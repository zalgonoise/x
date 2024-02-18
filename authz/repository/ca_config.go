package repository

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/authz/log"
	"github.com/zalgonoise/x/authz/metrics"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	defaultCleanupTimeout  = 5 * time.Minute
	defaultCleanupSchedule = "0 6 * * *" // every day at 6 AM
)

type Config struct {
	cleanupTimeout  time.Duration
	cleanupSchedule string

	logger *slog.Logger
	m      Metrics
	tracer trace.Tracer
}

func defaultConfig() Config {
	return Config{
		cleanupTimeout:  defaultCleanupTimeout,
		cleanupSchedule: defaultCleanupSchedule,
		logger:          log.NoOp(),
		m:               metrics.NoOp(),
		tracer:          noop.NewTracerProvider().Tracer("ca_cleanup"),
	}
}

func WithCleanupTimeout(dur time.Duration) cfg.Option[Config] {
	if dur <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.cleanupTimeout = dur

		return config
	})
}

func WithCleanupSchedule(cron string) cfg.Option[Config] {
	if cron == "" {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.cleanupSchedule = cron

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

func WithTrace(tracer trace.Tracer) cfg.Option[Config] {
	if tracer == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.tracer = tracer

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
