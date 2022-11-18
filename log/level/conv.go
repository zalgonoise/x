package level

var (
	lvKeys = map[string]Level{
		"trace": lTrace,
		"debug": lDebug,
		"info":  lInfo,
		"warn":  lWarn,
		"error": lError,
		"fatal": lFatal,
	}
	lvValues = map[lv]string{
		lTrace: "trace",
		lDebug: "debug",
		lInfo:  "info",
		lWarn:  "warn",
		lError: "error",
		lFatal: "fatal",
	}
)
