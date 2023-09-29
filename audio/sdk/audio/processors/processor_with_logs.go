package processors

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/zalgonoise/x/audio/sdk/audio"
)

type loggedProcessor struct {
	p      audio.Processor
	logger *slog.Logger
}

func (p loggedProcessor) Process(ctx context.Context, reader io.Reader) {
	p.logger.InfoContext(ctx, "processing signal from stream")

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
		p.logger.WarnContext(ctx, "error when shutting down", slog.String("error", err.Error()))

		return err
	}

	return nil
}

func ProcessorWithLogs(p audio.Processor, handler slog.Handler) audio.Processor {
	if p == nil {
		return audio.NoOpProcessor()
	}

	if handler == nil {
		handler = newDefaultHandler()
	}

	return loggedProcessor{
		p:      p,
		logger: slog.New(handler),
	}
}

func newDefaultHandler() slog.Handler {
	return slog.NewTextHandler(os.Stderr, nil)
}
