package log

import (
	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/level"
)

// Trace prints the input `msg` and `attrs` attributes as a Trace-level
// log message
func Trace(msg string, attrs ...attr.Attr) {
	std.Trace(msg, attrs...)
}

// Debug prints the input `msg` and `attrs` attributes as a Debug-level
// log message
func Debug(msg string, attrs ...attr.Attr) {
	std.Debug(msg, attrs...)
}

// Info prints the input `msg` and `attrs` attributes as a Info-level
// log message
func Info(msg string, attrs ...attr.Attr) {
	std.Info(msg, attrs...)
}

// Warn prints the input `msg` and `attrs` attributes as a Warn-level
// log message
func Warn(msg string, attrs ...attr.Attr) {
	std.Warn(msg, attrs...)
}

// Error prints the input `msg` and `attrs` attributes as a Error-level
// log message
func Error(msg string, attrs ...attr.Attr) {
	std.Error(msg, attrs...)
}

// Fatal prints the input `msg` and `attrs` attributes as a Fatal-level
// log message
func Fatal(msg string, attrs ...attr.Attr) {
	std.Fatal(msg, attrs...)
}

// Log prints the input `msg` and `attrs` attributes, as a log message
// with level `level`
func Log(level level.Level, msg string, attrs ...attr.Attr) {
	std.Log(level, msg, attrs...)
}

// SetDefault replaces this library's standard logger with `l`
func SetDefault(l Logger) {
	std = l
}
