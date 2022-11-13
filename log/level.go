package log

type lv int

type Level interface {
	String() string
	Int() int
}

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

func (l lv) String() string {
	return lvValues[l]
}

func (l lv) Int() int {
	return (int)(l)
}

func AsLevel(level string) Level {
	if l, ok := lvKeys[level]; ok {
		return l
	}
	return nil
}
