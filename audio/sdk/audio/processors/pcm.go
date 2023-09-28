package processors

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalgonoise/x/audio/errs"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters"
	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/header"
)

const (
	errDomain = errs.Domain("x/audio/sdk/processors/pcm")

	ErrHalt = errs.Kind("process stopped")

	ErrSignal = errs.Entity("with OS signal")
)

var ErrHaltSignal = errs.New(errDomain, ErrHalt, ErrSignal)

type pcm struct {
	exporter audio.Exporter

	errCh  chan error
	cancel context.CancelFunc
	stream *wav.Stream
}

func (p *pcm) Process(ctx context.Context, reader io.Reader) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer close(signalCh)

	ctx, p.cancel = context.WithCancel(ctx)

	go p.stream.Stream(ctx, reader, p.errCh)

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

func (p *pcm) Err() <-chan error {
	return p.errCh
}

func (p *pcm) ForceFlush() error {
	return p.exporter.ForceFlush()
}

func (p *pcm) Shutdown(ctx context.Context) error {
	defer close(p.errCh)
	p.cancel()

	return p.exporter.Shutdown(ctx)
}

func PCM(e ...audio.Exporter) audio.Processor {
	if len(e) == 0 {
		return audio.NoOpProcessor()
	}

	exporter := exporters.Multi(e...)

	return &pcm{
		exporter: exporter,
		errCh:    make(chan error),
		stream: wav.NewStream(nil, func(h *header.Header, data []float64) error {
			return exporter.Export(h, data)
		}),
	}
}
