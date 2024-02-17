package log

import (
	"context"
	"log/slog"
	"os"
)

func New(level string) *slog.Logger {
	var logLevel slog.Level

	err := logLevel.UnmarshalText([]byte(level))
	if err != nil {
		logLevel = slog.LevelInfo
	}

	logger := slog.New(
		NewSpanContextHandler(
			slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				AddSource: true,
				Level:     logLevel,
			}),
			true,
		),
	)

	if err != nil {
		logger.WarnContext(context.Background(), "invalid log level string",
			slog.String("input_level", level),
			slog.String("error", err.Error()),
		)
	}

	return logger
}
