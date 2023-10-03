package audio

import (
	"context"
	"errors"
)

type multiExporter struct {
	exporters []Exporter
}

// Export implements the Exporter interface.
//
// This call will iterate through all configured Exporters and return a wrapped error containing any raised errors
// from the Export call.
//
// This call is both blocking and sequential, as all Exporters are iterated through.
func (m multiExporter) Export(header Header, data []float64) error {
	errs := make([]error, 0, len(m.exporters))

	for i := range m.exporters {
		if err := m.exporters[i].Export(header, data); err != nil {
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
//
// The input parameter is variadic, and it will work with any number of items.
//
// If the caller does not pass any Exporters to the function, or they are all nil, a no-op Exporter is returned.
// If only one Exporter is passed to the function, it will be returned directly.
//
// For working with multiple exporters, a multiExporter type wraps a slice of Exporter. This function
// creates that data structure with a capacity based on the amount of Exporters passed in on input. Each item
// in the input is iterated on, where the function checks if it is nil (skipping it); and if it's also a
// multiExporter type. If it is a multiExporter, then all its (inner slice) Exporter items are appended to the returned
// multiExporter.
func MultiExporter(exporters ...Exporter) Exporter {
	switch len(exporters) {
	case 0:
		return NoOpExporter()
	case 1:
		if exporters[0] == nil {
			return NoOpExporter()
		}

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

		if len(me.exporters) == 0 {
			return NoOpExporter()
		}

		return me
	}
}
