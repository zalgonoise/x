package level

var (
	lvKeys = map[string]lv{
		"trace": Trace,
		"debug": Debug,
		"info":  Info,
		"warn":  Warn,
		"error": Error,
		"fatal": Fatal,
	}
	lvValues = map[lv]string{
		Trace: "trace",
		Debug: "debug",
		Info:  "info",
		Warn:  "warn",
		Error: "error",
		Fatal: "fatal",
	}
)
