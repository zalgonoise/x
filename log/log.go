package log

func Trace(msg string, attrs ...Attr)
func Debug(msg string, attrs ...Attr)
func Info(msg string, attrs ...Attr)
func Warn(msg string, attrs ...Attr)
func Error(msg string, attrs ...Attr)
func Fatal(msg string, attrs ...Attr)

func Log(level Level, msg string, attrs ...Attr)

func SetDefault(Logger)
