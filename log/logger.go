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
	// WithLevel will spawn a copy of this Logger with the input level `level`
	// as a verbosity filter
	WithLevel(level level.Level) Logger
}

var std = New(texth.New(os.Stderr))

type logger struct {
	h        handlers.Handler
	attrs    []attr.Attr
	levelRef level.Level
}

// New spawns a new logger based on the handler `h`
func New(h handlers.Handler) Logger {
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
		h:        std.Handler(),
		levelRef: (std).(*logger).levelRef,
		attrs:    attrs,
	}
}

// WithLevel will spawn a copy of this Logger with the input level `level`
// as a verbosity filter
func (l *logger) WithLevel(level level.Level) Logger {
	new := &logger{
		h:        l.h,
		levelRef: level,
	}
	copy(new.attrs, l.attrs)
	return new
}

// With will spawn a copy of this Logger with the input attributes
// `attrs`
func (l *logger) With(attrs ...attr.Attr) Logger {
	return &logger{
		h:        l.h,
		levelRef: l.levelRef,
		attrs:    attrs,
	}
}

// Enabled returns a boolean on whether the logger is accepting
// records with log level `level`
func (l *logger) Enabled(level level.Level) bool {
	if l.levelRef == nil || level.Int() >= l.levelRef.Int() {
		return true
	}
	return false
}

// Handler returns this Logger's Handler interface
func (l *logger) Handler() handlers.Handler {
	return l.h
}
