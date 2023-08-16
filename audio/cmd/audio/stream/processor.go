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

// ProcessorConfig enumerates general options that may or may not be used by a processor
type ProcessorConfig struct {
	// Size limits a certain size, in the context of the processor
	Size int
	// SampleRate denotes the input signal's sample rate
	SampleRate int
}

// ProcessorFunc is a type of function that creates other ProcessFunc. The resulting ProcessFunc
// will pipe the (processed) data to the input Exporter, using the ProcessorConfig if it needs so
type ProcessorFunc func(exporter Exporter, config *ProcessorConfig) ProcessFunc

// ProcessFunc is a function used by gbuf.RingFilter which is called on each pass through read data, as floating-point
// audio. It is always the responsibility of the caller or user to create any processing for that data
//
// An error can be returned from this function, indicating to the gbuf.RingFilter that reading should stop.
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

// PCMProcessor is a Processor type that processes an io.Reader as PCM audio
//
// Multiple ProcessorFunc are accepted, as to shape the ProcessFunc on both what to do with the audio stream
// and where to send it.
//
// An Exporter is also a part of the Processor, as one is directly linked to the other. The exported types
// are to allow direct interaction with the Exporter interface within the PCMProcessor, as part of the logic of the
// application.
//
// Finally, a ProcessorConfig can be added, which is passed into all input ProcessorFunc.
//
// TODO: replace Stream implementation with this type
type PCMProcessor struct {
	Exporter      Exporter
	processorFunc func([]float64) error

	cancel context.CancelFunc
	err    error
}

// Process consumes the input io.Reader as a WAV buffer, while also moving the
// processed data into the exporter.
//
// Process is a blocking operation but should be executed in a goroutine. If an operation fails
// during this call, Process and the underlying audio stream read are halted, and its internal error value
// (retrievable with Err()) is populated.
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

// Shutdown gracefully stops the processor, returning any stored error if it exists
func (p *PCMProcessor) Shutdown(_ context.Context) error {
	p.cancel()

	return p.err
}

// Err returns any internal error in the PCMProcessor, raised while reading the audio stream
func (p *PCMProcessor) Err() error {
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

// ProcessPeaks is a ProcessorFunc that processes the signal to extract the peak value (in intensity) in the signal
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

// ProcessSpectrum is a ProcessorFunc that processes the signal to extract the peak frequency (in magnitude, relative to
// the other analyzed frequencies) in the signal
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
