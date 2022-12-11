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

func (e loggerExporter) Export(trace spanner.Trace) {
	e.log.Trace(trace.ID().String(),
		attr.New("spans", trace.Extract()),
	)
}

func Logger(log logx.Logger) spanner.Exporter {
	return loggerExporter{
		log: log,
	}
}
