package stream

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalgonoise/x/audio/errs"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/wav"
)

// Exporter describes the actions supported by an audio metadata exporter
type Exporter interface {
	// SetPeakValue registers the float64 `data` value as an audio peak
	SetPeakValue(data float64) (err error)
	// ObserveFrequencies keeps track of changes in the registered frequencies
	ObserveFrequencies(frequencies []fft.FrequencyPower) (err error)
	// Shutdown gracefully stops the Exporter
	Shutdown(ctx context.Context) (err error)
}

type ProcessorConfig struct {
	Size       int
	SampleRate int
}

type ProcessorFunc func(exporter Exporter, config *ProcessorConfig) ProcessFunc
type ProcessFunc func(data []float64) error

const (
	processorDomain = errs.Domain("audio/stream/processor")

	ErrProcessorFunc = errs.Entity("processor functions")
	ErrExporter      = errs.Entity("exporter")
)

var (
	ErrEmptyProcessorFunc = errs.New(processorDomain, ErrEmpty, ErrProcessorFunc)
	ErrEmptyExporter      = errs.New(processorDomain, ErrEmpty, ErrExporter)
)

type PCMProcessor struct {
	Exporter      Exporter
	processorFunc func([]float64) error

	cancel context.CancelFunc
	err    error
}

func (p *PCMProcessor) Process(ctx context.Context, reader io.Reader) {
	var (
		signalCh = make(chan os.Signal, 1)
		errCh    = make(chan error)
		stream   = wav.NewStream(nil, p.processorFunc)
	)

	ctx, cancel := context.WithCancel(ctx)

	p.cancel = func() {
		cancel()
		if closer, ok := (reader).(io.Closer); ok {
			_ = closer.Close()
		}

		close(errCh)
		close(signalCh)
	}

	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM)

	go stream.Stream(ctx, reader, errCh)

	for {
		select {
		case <-ctx.Done():
			p.err = ctx.Err()

			return
		case <-signalCh:
			p.err = ctx.Err()

			return
		case err := <-errCh:
			if !errors.Is(err, context.DeadlineExceeded) {
				p.err = err
			}

			return
		}
	}
}

func (p *PCMProcessor) Shutdown(_ context.Context) error {
	p.cancel()

	return p.err
}

func NewPCMProcessor(
	exporter Exporter,
	config *ProcessorConfig,
	processorFuncs ...ProcessorFunc,
) (*PCMProcessor, error) {
	if exporter == nil {
		return nil, ErrEmptyExporter
	}

	if len(processorFuncs) == 0 {
		return nil, ErrEmptyProcessorFunc
	}

	processors := make([]func([]float64) error, 0, len(processorFuncs))
	for i := range processorFuncs {
		processors = append(processors, processorFuncs[i](exporter, config))
	}

	return &PCMProcessor{
		processorFunc: wav.MultiProc(false, processors...),
	}, nil
}

func ProcessPeaks(exporter Exporter, _ *ProcessorConfig) ProcessFunc {
	return func(data []float64) error {
		var maximum float64

		for i := range data {
			if data[i] > maximum {
				maximum = data[i]
			}
		}

		return exporter.SetPeakValue(maximum)
	}
}

func ProcessSpectrum(exporter Exporter, config *ProcessorConfig) ProcessFunc {
	bs := fft.NearestBlock(config.Size)
	windowBlock := window.New(window.Blackman, int(bs))

	return func(data []float64) error {
		for i := 0; i+int(bs) < len(data); i += int(bs) {
			spectrum := fft.Apply(
				config.SampleRate,
				data[i:i+int(bs)],
				windowBlock,
			)

			if err := exporter.ObserveFrequencies(spectrum); err != nil {
				return err
			}
		}

		return nil
	}
}
