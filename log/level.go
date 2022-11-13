package log

type lv int

const (
	LTrace lv = iota
	LDebug
	LInfo
	LWarn
	LError
	LFatal
)

var (
	lvKeys = map[string]lv{
		"trace": LTrace,
		"debug": LDebug,
		"info":  LInfo,
		"warn":  LWarn,
		"error": LError,
		"fatal": LFatal,
	}
	lvValues = map[lv]string{
		LTrace: "trace",
		LDebug: "debug",
		LInfo:  "info",
		LWarn:  "warn",
		LError: "error",
		LFatal: "fatal",
	}
)

type Level interface {
	String() string
}

func (l lv) String() string {
	return lvValues[l]
}

func AsLevel(level string) Level {
	if l, ok := lvKeys[level]; ok {
		return l
	}
	return nil
}
