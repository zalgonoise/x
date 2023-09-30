package audio

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalgonoise/x/audio/errs"
)

const (
	ErrHalt = errs.Kind("process stopped")

	ErrSignal = errs.Entity("with OS signal")
)

var ErrHaltSignal = errs.New(errDomain, ErrHalt, ErrSignal)

type processor struct {
	streamExporter StreamExporter

	errCh  chan error
	cancel context.CancelFunc
}

func (p *processor) Process(ctx context.Context, reader io.Reader) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer close(signalCh)

	ctx, p.cancel = context.WithCancel(ctx)

	go p.streamExporter.Stream(ctx, reader, p.errCh)

	for {
		select {
		case <-ctx.Done():
			p.cancel()

			return
		case sig := <-signalCh:
			p.errCh <- fmt.Errorf("%w: %s", ErrHaltSignal, sig.String())
			p.cancel()

			return
		}
	}
}

func (p *processor) Err() <-chan error {
	return p.errCh
}

func (p *processor) ForceFlush() error {
	return p.streamExporter.ForceFlush()
}

func (p *processor) Shutdown(ctx context.Context) error {
	defer close(p.errCh)
	p.cancel()

	return p.streamExporter.Shutdown(ctx)
}

func NewProcessor(streamExporter StreamExporter) Processor {
	if streamExporter == nil {
		return NoOpProcessor()
	}

	return &processor{
		errCh:          make(chan error),
		streamExporter: streamExporter,
	}
}
