package level

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
