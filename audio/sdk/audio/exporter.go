package audio

import (
	"context"
	"errors"

	"github.com/zalgonoise/x/audio/wav/header"
)

// Exporter is responsible for pushing the processed data into a certain destination.
//
// It is of the implementer's responsibility to comply with any requirements needed to export the audio,
// which is out-of-scope for this package.
//
// Exporter will periodically receive data either from the Processor or its Collector functions. The data
// is ready to be pushed to its destination by the time it is passed to the Exporter and there shouldn't be any
// further processing.
//
// Its Export method will propagate the incoming audio data chunk into its aggregators if configured, or directly to
// its output target if none are configured.
//
// Exporter also implements StreamCloser as a means to both flush any batched or aggregated values and gracefully
// shutdown the exporter
type Exporter interface {
	// Export consumes the audio data chunks from the Processor, preparing them to be pushed to their destination.
	//
	// The Exporter may have a set of Collector configured -- in that case it send the audio data it receives to all
	// Collector. It is also the responsibility of the implementation of Exporter to properly send those aggregations
	// or batches to their correct destination once they're done being collected, or on a frequent flush.
	//
	// The returned error from an Export call is related to an error raised when pushing the values or items to the
	// target, or from any errors raised by the configured Collector types.
	Export(h *header.Header, data []float64) error
	// StreamCloser defines common methods when interacting with a streaming module, targeting actions to either flush
	// the module or to shut it down gracefully.
	StreamCloser
}

type noOpExporter struct{}

// Export implements the Exporter interface.
//
// This is a no-op call and the returned error is always nil.
func (noOpExporter) Export(*header.Header, []float64) error { return nil }

// ForceFlush implements the Exporter and StreamCloser interfaces.
//
// This is a no-op call and the returned error is always nil.
func (noOpExporter) ForceFlush() error { return nil }

// Shutdown implements the Exporter, Closer and StreamCloser interfaces.
//
// This is a no-op call and the returned error is always nil.
func (noOpExporter) Shutdown(context.Context) error { return nil }

// NoOpExporter returns a no-op Exporter
func NoOpExporter() Exporter {
	return noOpExporter{}
}

type multiExporter struct {
	exporters []Exporter
}

// Export implements the Exporter interface.
//
// This call will iterate through all configured Exporters and return a wrapped error containing any raised errors
// from the Export call.
//
// This call is both blocking and sequential, as all Exporters are iterated through.
func (m multiExporter) Export(h *header.Header, data []float64) error {
	errs := make([]error, 0, len(m.exporters))

	for i := range m.exporters {
		if err := m.exporters[i].Export(h, data); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// ForceFlush implements the Exporter interface.
//
// This call will iterate through all configured Exporters and return a wrapped error containing any raised errors
// from the ForceFlush call.
//
// This call is both blocking and sequential, as all Exporters are iterated through.
func (m multiExporter) ForceFlush() error {
	errs := make([]error, 0, len(m.exporters))

	for i := range m.exporters {
		if err := m.exporters[i].ForceFlush(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Shutdown implements the Exporter interface.
//
// This call will iterate through all configured Exporters and return a wrapped error containing any raised errors
// from the Shutdown call.
//
// This call is both blocking and sequential, as all Exporters are iterated through.
func (m multiExporter) Shutdown(ctx context.Context) error {
	errs := make([]error, 0, len(m.exporters))

	for i := range m.exporters {
		if err := m.exporters[i].Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// MultiExporter joins several Exporter interfaces as one, to facilitate its access when implementing
// Processor logic, without much repetition.
func MultiExporter(exporters ...Exporter) Exporter {
	switch len(exporters) {
	case 0:
		return NoOpExporter()
	case 1:
		return exporters[0]
	default:
		me := multiExporter{
			exporters: make([]Exporter, 0, len(exporters)),
		}

		for i := range exporters {
			switch v := exporters[i].(type) {
			case nil:
				continue
			case multiExporter:
				me.exporters = append(me.exporters, v.exporters...)
			default:
				me.exporters = append(me.exporters, v)
			}
		}

		return me
	}
}
