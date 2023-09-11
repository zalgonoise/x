package log

import (
	"log/slog"

	"github.com/zalgonoise/x/cfg"
)

const (
	formatJSON = iota
	formatText
)

type Config struct {
	handler slog.Handler

	format       int
	noSource     bool
	level        slog.Leveler
	replaceAttrs func(groups []string, a slog.Attr) slog.Attr

	withTraceID bool
	withSpanID  bool
}

func WithHandler(h slog.Handler) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.handler = h

		return config
	})
}

func AsText() cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.format = formatText

		return config
	})
}

func AsJSON() cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.format = formatJSON

		return config
	})
}

func WithoutSource() cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.noSource = true

		return config
	})
}

func WithLevel(level slog.Leveler) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.level = level

		return config
	})
}

func WithReplaceAttrs(fn func(groups []string, a slog.Attr) slog.Attr) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.replaceAttrs = fn

		return config
	})
}

func WithTraceContext(withTraceID, withSpanID bool) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.withTraceID = withTraceID
		config.withSpanID = withSpanID

		return config
	})
}
