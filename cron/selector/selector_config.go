package selector

import (
	"log/slog"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/cron/executor"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	exec []executor.Executor

	handler slog.Handler
	metrics Metrics
	tracer  trace.Tracer
}

// WithExecutors configures the Selector with the input executor.Executor(s).
//
// This call returns a cfg.NoOp cfg.Option if the input set of executor.Executor is empty, or contains
// only nil and / or no-op executor.Executor.
func WithExecutors(executors ...executor.Executor) cfg.Option[Config] {
	execs := make([]executor.Executor, 0, len(executors))
	for i := range executors {
		if executors[i] == nil || executors[i] == executor.NoOp() {
			continue
		}

		execs = append(execs, executors[i])
	}

	if len(execs) == 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		if len(config.exec) == 0 {
			config.exec = execs

			return config
		}

		config.exec = append(config.exec, execs...)

		return config
	})
}

// WithMetrics decorates the Selector with the input metrics registry.
func WithMetrics(m Metrics) cfg.Option[Config] {
	if m == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.metrics = m

		return config
	})
}

// WithLogger decorates the Selector with the input logger.
func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	if logger == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.handler = logger.Handler()

		return config
	})
}

// WithLogHandler decorates the Selector with logging using the input log handler.
func WithLogHandler(handler slog.Handler) cfg.Option[Config] {
	if handler == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.handler = handler

		return config
	})
}

// WithTrace decorates the Selector with the input trace.Tracer.
func WithTrace(tracer trace.Tracer) cfg.Option[Config] {
	if tracer == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.tracer = tracer

		return config
	})
}
