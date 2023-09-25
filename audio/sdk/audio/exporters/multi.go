package exporters

import (
	"context"
	"errors"

	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/wav/header"
)

type multiExporter struct {
	exporters []audio.Exporter
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

// Multi joins several Exporter interfaces as one, to facilitate its access when implementing
// Processor logic, without much repetition.
func Multi(exporters ...audio.Exporter) audio.Exporter {
	switch len(exporters) {
	case 0:
		return audio.NoOpExporter()
	case 1:
		return exporters[0]
	default:
		me := multiExporter{
			exporters: make([]audio.Exporter, 0, len(exporters)),
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
