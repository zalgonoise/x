package selector

import (
	"context"
	"log/slog"
	"os"
)

type withLogs struct {
	s      Selector
	logger *slog.Logger
}

func (s withLogs) Next(ctx context.Context) error {
	s.logger.InfoContext(ctx, "selecting the next task")

	if err := s.s.Next(ctx); err != nil {
		s.logger.ErrorContext(ctx, "failed to select and execute the next task", slog.String("error", err.Error()))

		return err
	}

	return nil
}

func selectorWithLogs(s Selector, handler slog.Handler) Selector {
	if s == nil {
		return noOpSelector{}
	}

	if handler == nil {
		handler = slog.NewTextHandler(os.Stderr, nil)
	}

	if logs, ok := s.(withLogs); ok {
		logs.logger = slog.New(handler)

		return logs
	}

	return withLogs{
		s:      s,
		logger: slog.New(handler),
	}
}
