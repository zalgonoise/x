package log

import (
	"time"

	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/level"
	"github.com/zalgonoise/x/log/records"
)

// Printer interface describes the behavior that a (log) Printer
// should have. This includes individual methods for printing log
// messages for each log level, as well as a general-purpose `Log()`
// method to customize the log level.
type Printer interface {
	// Trace prints a log message `msg` with attributes `attrs`, with
	// Trace-level
	Trace(msg string, attrs ...attr.Attr)
	// Debug prints a log message `msg` with attributes `attrs`, with
	// Debug-level
	Debug(msg string, attrs ...attr.Attr)
	// Info prints a log message `msg` with attributes `attrs`, with
	// Info-level
	Info(msg string, attrs ...attr.Attr)
	// Warn prints a log message `msg` with attributes `attrs`, with
	// Warn-level
	Warn(msg string, attrs ...attr.Attr)
	// Error prints a log message `msg` with attributes `attrs`, with
	// Error-level
	Error(msg string, attrs ...attr.Attr)
	// Fatal prints a log message `msg` with attributes `attrs`, with
	// Fatal-level
	Fatal(msg string, attrs ...attr.Attr)
	// Log prints a log message `msg` with attributes `attrs`, with
	// `level` log level
	Log(level level.Level, msg string, attrs ...attr.Attr)
}

// Log prints a log message `msg` with attributes `attrs`, with
// `level` log level
func (l *logger) Log(lv level.Level, msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}
	if lv == nil {
		lv = level.Info
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), lv, msg, rAttr...)
	_ = l.h.Handle(r)
}

// Trace prints a log message `msg` with attributes `attrs`, with
// Trace-level
func (l *logger) Trace(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Trace, msg, rAttr...)
	_ = l.h.Handle(r)
}

// Debug prints a log message `msg` with attributes `attrs`, with
// Debug-level
func (l *logger) Debug(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Debug, msg, rAttr...)
	_ = l.h.Handle(r)
}

// Info prints a log message `msg` with attributes `attrs`, with
// Info-level
func (l *logger) Info(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Info, msg, rAttr...)
	_ = l.h.Handle(r)
}

// Warn prints a log message `msg` with attributes `attrs`, with
// Warn-level
func (l *logger) Warn(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Warn, msg, rAttr...)
	_ = l.h.Handle(r)
}

// Error prints a log message `msg` with attributes `attrs`, with
// Error-level
func (l *logger) Error(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Error, msg, rAttr...)
	_ = l.h.Handle(r)
}

// Fatal prints a log message `msg` with attributes `attrs`, with
// Fatal-level
func (l *logger) Fatal(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Fatal, msg, rAttr...)
	_ = l.h.Handle(r)
}
