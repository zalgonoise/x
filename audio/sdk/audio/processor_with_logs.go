package audio

import (
	"context"
	"io"
	"log/slog"
)

type loggedProcessor struct {
	p      Processor
	logger *slog.Logger
}

func (p loggedProcessor) Process(ctx context.Context, reader io.Reader) {
	p.logger.InfoContext(ctx, "processing signal from streamer")

	p.p.Process(ctx, reader)

	p.logger.InfoContext(ctx, "terminating Process call")
}

func (p loggedProcessor) Err() <-chan error {
	return p.p.Err()
}

func (p loggedProcessor) ForceFlush() error {
	ctx := context.Background()
	p.logger.InfoContext(ctx, "flushing")

	if err := p.p.ForceFlush(); err != nil {
		p.logger.WarnContext(ctx, "error when flushing", slog.String("error", err.Error()))

		return err
	}

	return nil
}

func (p loggedProcessor) Shutdown(ctx context.Context) error {
	p.logger.InfoContext(ctx, "shutting down Processor")

	if err := p.p.Shutdown(ctx); err != nil {
		p.logger.WarnContext(ctx, "failed to gracefully shut down", slog.String("error", err.Error()))

		return err
	}

	return nil
}

func ProcessorWithLogs(p Processor, handler slog.Handler) Processor {
	if p == nil {
		return NoOpProcessor()
	}

	if handler == nil {
		handler = newDefaultHandler()
	}

	if withLogs, ok := (p).(loggedProcessor); ok {
		withLogs.logger = slog.New(handler)

		return withLogs
	}

	return loggedProcessor{
		p:      p,
		logger: slog.New(handler),
	}
}
