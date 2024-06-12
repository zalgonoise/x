package actions

import (
	"context"

	"github.com/zalgonoise/x/modupdate/events"
)

type noOpReporter struct{}

func (noOpReporter) ReportEvent(context.Context, events.Event) {}
