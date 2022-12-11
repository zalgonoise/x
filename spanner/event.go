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
	return &event{
		name:      name,
		timestamp: time.Now(),
		attrs:     attrs,
	}
}
