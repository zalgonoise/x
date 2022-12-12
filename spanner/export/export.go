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

func (e loggerExporter) Export(ctx context.Context, spans []spanner.SpanData) error {
	e.log.Trace("spanner",
		attr.New("spans", spans),
	)
	return nil
}
func (e loggerExporter) Shutdown(ctx context.Context) error {
	return nil
}

func Logger(log logx.Logger) spanner.Exporter {
	return loggerExporter{
		log: log,
	}
}
