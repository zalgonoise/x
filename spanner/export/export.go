package export

import (
	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/x/spanner"
)

type loggerExporter struct {
	log interface {
		Trace(msg string, attrs ...attr.Attr)
	}
}

func (e loggerExporter) Export(traceID spanner.TraceID, spans ...spanner.SpanData) error {
	e.log.Trace(traceID.String(),
		attr.New("spans", spans),
	)
	return nil
}

func Logger(log logx.Logger) spanner.Exporter {
	return loggerExporter{
		log: log,
	}
}
