package log

import (
	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/level"
)

func Trace(msg string, attrs ...attr.Attr) {
	std.Trace(msg, attrs...)
}
func Debug(msg string, attrs ...attr.Attr) {
	std.Debug(msg, attrs...)
}
func Info(msg string, attrs ...attr.Attr) {
	std.Info(msg, attrs...)
}
func Warn(msg string, attrs ...attr.Attr) {
	std.Warn(msg, attrs...)
}
func Error(msg string, attrs ...attr.Attr) {
	std.Error(msg, attrs...)
}
func Fatal(msg string, attrs ...attr.Attr) {
	std.Fatal(msg, attrs...)
}

func Log(level level.Level, msg string, attrs ...attr.Attr) {
	std.Log(level, msg, attrs...)
}

func SetDefault(l Logger) {
	std = l
}
