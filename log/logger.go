package log

import (
	"os"

	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/handlers"
	"github.com/zalgonoise/x/log/handlers/texth"
	"github.com/zalgonoise/x/log/level"
)

// Logger interface describes the behavior that a logger should
// have
//
// This includes the Printer interface, as well as other methods
// to give the logger more flexibility
type Logger interface {
	// Printer interface allows registering log messages
	Printer
	// Enabled returns a boolean on whether the logger is accepting
	// records with log level `level`
	Enabled(level level.Level) bool
	// Handler returns this Logger's Handler interface
	Handler() handlers.Handler
	// With will spawn a copy of this Logger with the input attributes
	// `attrs`
	With(attrs ...attr.Attr) Logger
}

var std = New(texth.New(os.Stderr))

type logger struct {
	h     handlers.Handler
	attrs []attr.Attr
}

// New spawns a new logger based on the handler `h`
func New(h handlers.Handler) Logger {
	if h == nil {
		return nil
	}
	return &logger{
		h: h,
	}
}

// Default returns the standard logger for this library
func Default() Logger {
	return std
}

// With will spawn a copy of this library's standard Logger
// with the input attributes `attrs`
func With(attrs ...attr.Attr) Logger {
	return &logger{
		h:     std.Handler(),
		attrs: attrs,
	}
}

// With will spawn a copy of this Logger with the input attributes
// `attrs`
func (l *logger) With(attrs ...attr.Attr) Logger {
	return &logger{
		h:     l.h,
		attrs: attrs,
	}
}

// Enabled returns a boolean on whether the logger is accepting
// records with log level `level`
func (l *logger) Enabled(level level.Level) bool {
	if level == nil {
		return true
	}
	return l.h.Enabled(level)
}

// Handler returns this Logger's Handler interface
func (l *logger) Handler() handlers.Handler {
	return l.h
}
