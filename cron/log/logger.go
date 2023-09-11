package log

import (
	"log/slog"
	"os"

	"github.com/zalgonoise/x/cfg"
)

func New(options ...cfg.Option[Config]) *slog.Logger {
	config := cfg.New(options...)

	handler := newHandler(config)

	if config.withTraceID {
		handler = NewSpanContextHandler(handler, config.withSpanID)
	}

	return slog.New(handler)
}

func newHandler(config Config) slog.Handler {
	if config.handler != nil && config.handler != slog.Handler(nil) {
		return config.handler
	}

	addSource := true
	if config.noSource {
		addSource = false
	}

	level := config.level
	if level == nil {
		level = slog.LevelInfo
	}

	var replaceAttrs func(groups []string, a slog.Attr) slog.Attr
	if config.replaceAttrs != nil {
		replaceAttrs = config.replaceAttrs
	}

	switch config.format {
	case formatText:
		return slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource:   addSource,
			Level:       level,
			ReplaceAttr: replaceAttrs,
		})
	default:
		return slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			AddSource:   addSource,
			Level:       level,
			ReplaceAttr: replaceAttrs,
		})
	}
}
