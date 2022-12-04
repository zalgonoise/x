package spanner

import (
	"time"

	"github.com/zalgonoise/logx/attr"
)

type event struct {
	name      string
	timestamp time.Time
	attrs     []attr.Attr
}

func newEvent(name string, attrs ...attr.Attr) event {
	return event{
		name:      name,
		timestamp: time.Now(),
		attrs:     attrs,
	}
}

func (e event) Export() EventData {
	return EventData{
		Name:       e.name,
		Timestamp:  e.timestamp,
		Attributes: e.attrs,
	}
}
