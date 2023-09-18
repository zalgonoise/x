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

	chunks chan dataChunk
	errCh  chan error
	cancel context.CancelFunc
	stream *wav.Stream
}

type dataChunk struct {
	h    *header.Header
	data []float64
}

func (p *pcm) Process(ctx context.Context, reader io.Reader) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM)

	ctx, p.cancel = context.WithCancel(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case chunk := <-p.chunks:
				for i := range p.exporters {
					if err := p.exporters[i].Export(chunk.h, chunk.data); err != nil {
						p.errCh <- err

						p.Shutdown(context.Background())
					}

					return
				}
			}
		}
	}()

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

	close(p.chunks)
	close(p.errCh)

	return errors.Join(errs...)
}

func NewPCM(exporters ...audio.Exporter) audio.Processor {
	if len(exporters) == 0 {
		return nil
	}

	chunks := make(chan dataChunk)

	return &pcm{
		exporters: exporters,
		errCh:     make(chan error),
		chunks:    chunks,
		stream: wav.NewStream(nil, func(h *header.Header, data []float64) error {
			chunks <- dataChunk{
				h:    h,
				data: data,
			}

			return nil
		}),
	}
}
