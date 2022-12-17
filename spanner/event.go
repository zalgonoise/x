package spanner

import (
	"time"

	"github.com/zalgonoise/attr"
)

type event struct {
	name      string
	timestamp time.Time
	attrs     []attr.Attr
}

func newEvent(name string, attrs ...attr.Attr) *event {
	e := &event{
		name:      name,
		timestamp: time.Now(),
	}

	if len(attrs) > 0 {
		e.attrs = attrs
	}

	return e
}
