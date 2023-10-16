package selector

import (
	"context"
	"log/slog"
	"os"
)

type SelectorWithLogs struct {
	s      Selector
	logger *slog.Logger
}

func (s SelectorWithLogs) Next(ctx context.Context) error {
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

	if withLogs, ok := s.(SelectorWithLogs); ok {
		withLogs.logger = slog.New(handler)

		return withLogs
	}

	return SelectorWithLogs{
		s:      s,
		logger: slog.New(handler),
	}
}
