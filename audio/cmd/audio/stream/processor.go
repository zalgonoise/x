package stream

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalgonoise/x/audio/errs"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/header"
)

// Exporter describes the actions supported by an audio metadata exporter
type Exporter interface {
	// SendPeak registers the float64 `data` value as an audio peak in the exporter
	SendPeak(data float64) (err error)
	// SendSpectrum registers the frequency spectrum in the exporter
	SendSpectrum(frequencies []fft.FrequencyPower) (err error)
	// ForceFlush is used in Exporter implementations that buffer or batch their values, as a means of immediately
	// exporting any values that are in-memory.
	ForceFlush(ctx context.Context) (err error)
	// Shutdown gracefully stops the Exporter
	Shutdown(ctx context.Context) (err error)
}

// ProcessorConfig enumerates general options that may or may not be used by a processor
type ProcessorConfig struct {
	// Size limits a certain size, in the context of the processor
	Size int
}

// ProcessorFunc is a type of function that creates other ProcessFunc. The resulting wav.ProcessFunc
// will pipe the (processed) data to the input Exporter, using the ProcessorConfig if it needs so
type ProcessorFunc func(exporter Exporter, config *ProcessorConfig) wav.ProcessFunc

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
	Exporter Exporter

	stream *wav.Stream
}

// Process consumes the input io.Reader as a WAV buffer, while also moving the
// processed data into the exporter.
//
// Process is a blocking operation but should be executed in a goroutine. If an operation fails
// during this call, Process and the underlying audio stream read are halted, and its internal error value
// (retrievable with Err()) is populated.
func (p *PCMProcessor) Process(ctx context.Context, cancel context.CancelCauseFunc, reader io.Reader) {
	var (
		err      = new(error)
		signalCh = make(chan os.Signal, 1)
		errorCh  = make(chan error)
	)

	defer func(err *error) {
		cancel(*err)
		if closer, ok := (reader).(io.Closer); ok {
			_ = closer.Close()
		}

		close(signalCh)
	}(err)

	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM)

	go p.stream.Stream(ctx, reader, errorCh)

	for {
		select {
		case procErr := <-errorCh:
			if procErr != nil {
				*err = procErr

				return
			}
		case <-ctx.Done():
			if ctxErr := ctx.Err(); ctxErr != nil {
				*err = ctxErr
			}

			return
		case sig := <-signalCh:
			*err = fmt.Errorf("received shutdown signal: %s", sig.String())

			return
		}
	}
}

// Shutdown gracefully stops the processor
func (p *PCMProcessor) Shutdown(_ context.Context) error {
	return nil
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

	processors := make([]wav.ProcessFunc, 0, len(processorFuncs))
	for i := range processorFuncs {
		processors = append(processors, processorFuncs[i](exporter, config))
	}

	return &PCMProcessor{
		stream: wav.NewStream(nil, wav.MultiProc(false, processors...)),
	}, nil
}

// ProcessPeaks is a ProcessorFunc that processes the signal to extract the peak value (in intensity) in the signal
func ProcessPeaks(exporter Exporter, _ *ProcessorConfig) wav.ProcessFunc {
	return func(h *header.Header, data []float64) error {
		var maximum float64

		for i := range data {
			if data[i] > maximum {
				maximum = data[i]
			}
		}

		return exporter.SendPeak(maximum)
	}
}

// ProcessSpectrum is a ProcessorFunc that processes the signal to extract the peak frequency (in magnitude, relative to
// the other analyzed frequencies) in the signal
func ProcessSpectrum(exporter Exporter, config *ProcessorConfig) wav.ProcessFunc {
	size := 64
	if config != nil && config.Size >= 8 {
		size = config.Size
	}

	sampleRate := 44100

	return func(h *header.Header, data []float64) error {
		if h != nil {
			sampleRate = int(h.SampleRate)
		}

		bs := fft.NearestBlock(size)
		windowBlock := window.New(window.Blackman, int(bs))

		for i := 0; i+int(bs) < len(data); i += int(bs) {
			spectrum := fft.Apply(
				sampleRate,
				data[i:i+int(bs)],
				windowBlock,
			)

			if err := exporter.SendSpectrum(spectrum); err != nil {
				return err
			}
		}

		return nil
	}
}
