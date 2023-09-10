package selector

import (
	"context"
	"log/slog"
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

func selectorWithLogs(s Selector, logger *slog.Logger) Selector {
	if s == nil {
		return noOpSelector{}
	}

	if logger == nil {
		return s
	}

	return SelectorWithLogs{
		s:      s,
		logger: logger,
	}
}
