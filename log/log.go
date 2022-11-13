package log

func Trace(msg string, attrs ...Attr) {
	stdLogger.Trace(msg, attrs...)
}
func Debug(msg string, attrs ...Attr) {
	stdLogger.Debug(msg, attrs...)
}
func Info(msg string, attrs ...Attr) {
	stdLogger.Info(msg, attrs...)
}
func Warn(msg string, attrs ...Attr) {
	stdLogger.Warn(msg, attrs...)
}
func Error(msg string, attrs ...Attr) {
	stdLogger.Error(msg, attrs...)
}
func Fatal(msg string, attrs ...Attr) {
	stdLogger.Fatal(msg, attrs...)
}

func Log(level Level, msg string, attrs ...Attr) {
	stdLogger.Log(level, msg, attrs...)
}

func SetDefault(l Logger) {
	stdLogger = l
}
