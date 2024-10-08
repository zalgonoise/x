package audio

import (
	"context"

	"github.com/zalgonoise/x/audio/encoding/wav"
)

type noOpExporter struct{}

// Export implements the Exporter interface.
//
// This is a no-op call and the returned error is always nil.
func (noOpExporter) Export(context.Context, *wav.Header, []float64) error { return nil }

// ForceFlush implements the Exporter and StreamCloser interfaces.
//
// This is a no-op call and the returned error is always nil.
func (noOpExporter) ForceFlush() error { return nil }

// Shutdown implements the Exporter, Closer and StreamCloser interfaces.
//
// This is a no-op call and the returned error is always nil.
func (noOpExporter) Shutdown(context.Context) error { return nil }

// NoOpExporter returns a no-op Exporter.
func NoOpExporter() Exporter {
	return noOpExporter{}
}
