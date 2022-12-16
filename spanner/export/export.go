package export

import (
	"context"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/x/spanner"
)

type loggerExporter struct {
	log interface {
		Trace(msg string, attrs ...attr.Attr)
	}
}

// Export pushes the input SpanData `spans` to its output, as a non-blocking
// function
func (e loggerExporter) Export(ctx context.Context, spans []spanner.SpanData) error {
	e.log.Trace("spanner",
		attr.New("spans", spans),
	)
	return nil
}

// Shutdown gracefully terminates the Exporter
func (e loggerExporter) Shutdown(ctx context.Context) error {
	return nil
}

// Logger returns an Exporter that is configured to write SpanData through
// a logx.Logger
func Logger(log logx.Logger) spanner.Exporter {
	return loggerExporter{
		log: log,
	}
}
