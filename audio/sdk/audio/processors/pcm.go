package processors

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/wav"
)

type pcm struct {
	exporter audio.Exporter

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
	return p.exporter.ForceFlush()
}

func (p *pcm) Shutdown(ctx context.Context) error {
	p.cancel()
	close(p.errCh)

	return p.exporter.Shutdown(ctx)
}

func NewPCM(exporters ...audio.Exporter) audio.Processor {
	if len(exporters) == 0 {
		return audio.NoOpProcessor()
	}

	exporter := audio.MultiExporter(exporters...)

	return &pcm{
		exporter: exporter,
		errCh:    make(chan error),
		stream:   wav.NewStream(nil, exporter.Export),
	}
}
