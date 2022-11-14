package log

import (
	"os"

	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/handlers"
	"github.com/zalgonoise/x/log/handlers/jsonh"
	"github.com/zalgonoise/x/log/level"
)

type Logger interface {
	Printer
	Enabled(level level.Level) bool
	Handler() handlers.Handler
	With(attrs ...attr.Attr) Logger
}

var std = New(jsonh.New(os.Stderr))

type logger struct {
	h        handlers.Handler
	attrs    []attr.Attr
	levelRef level.Level
}

func New(h handlers.Handler) Logger {
	return &logger{
		h: h,
	}
}

func Default() Logger {
	return std
}

func With(attrs ...attr.Attr) Logger {
	return &logger{
		h:        std.Handler(),
		levelRef: (std).(*logger).levelRef,
		attrs:    attrs,
	}
}

func (l *logger) WithLevel(level level.Level) Logger {
	new := &logger{
		h:        l.h,
		levelRef: level,
	}
	copy(new.attrs, l.attrs)
	return new
}

func (l *logger) With(attrs ...attr.Attr) Logger {
	return &logger{
		h:        l.h,
		levelRef: l.levelRef,
		attrs:    attrs,
	}
}

func (l *logger) Enabled(level level.Level) bool {
	if l.levelRef == nil || level.Int() >= l.levelRef.Int() {
		return true
	}
	return false
}
func (l *logger) Handler() handlers.Handler {
	return l.h
}
