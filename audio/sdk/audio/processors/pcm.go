package processors

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/header"
)

type pcm struct {
	exporters []audio.Exporter

	errCh  chan error
	cancel context.CancelFunc
	stream *wav.Stream
}

func (p *pcm) Process(ctx context.Context, reader io.Reader) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM)

	ctx, p.cancel = context.WithCancel(ctx)

	go p.stream.Stream(ctx, reader, p.errCh)

	for {
		select {
		case <-ctx.Done():
			p.cancel()

			return
		case sig := <-signalCh:
			p.errCh <- fmt.Errorf("received signal %s", sig.String())
			p.cancel()

			return
		}
	}
}

func (p *pcm) Err() <-chan error {
	return p.errCh
}

func (p *pcm) ForceFlush() error {
	errs := make([]error, 0, len(p.exporters))

	for i := range p.exporters {
		if err := p.exporters[i].ForceFlush(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (p *pcm) Shutdown(ctx context.Context) error {
	p.cancel()

	errs := make([]error, 0, len(p.exporters))

	for i := range p.exporters {
		if err := p.exporters[i].Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	close(p.errCh)

	return errors.Join(errs...)
}

func NewPCM(exporters ...audio.Exporter) audio.Processor {
	if len(exporters) == 0 {
		return nil
	}

	return &pcm{
		exporters: exporters,
		errCh:     make(chan error),
		stream: wav.NewStream(nil, func(h *header.Header, data []float64) error {
			errs := make([]error, 0, len(exporters))

			for i := range exporters {
				if err := exporters[i].Export(h, data); err != nil {
					errs = append(errs, err)
				}
			}

			return errors.Join(errs...)
		}),
	}
}
