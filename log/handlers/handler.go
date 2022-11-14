package handlers

import (
	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/level"
	"github.com/zalgonoise/x/log/records"
)

type Handler interface {
	Enabled(level level.Level) bool
	Handle(records.Record) error
	With(attrs ...attr.Attr) Handler
}
