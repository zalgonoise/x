package stream

import (
	"time"
)

// OptionFunc is a function type that implements the Option interface
//
// Similar to the http.Handler and http.HandleFunc in the standard library,
// the Apply method implementation will simply be the type itself (a function)
// calling the input Config as an argument
type OptionFunc func(c *Config)

// Apply implements the Option interface
func (o OptionFunc) Apply(c *Config) {
	o(c)
}

// MultiOption merges a slice of OptionFunc into one
func MultiOption(opts ...OptionFunc) OptionFunc {
	if len(opts) == 0 {
		return nil
	}
	mo := make([]OptionFunc, 0, len(opts))
	for i := range opts {
		if opts[i] == nil {
			continue
		}
		mo = append(mo, opts[i])
	}

	return func(c *Config) {
		for i := range mo {
			mo[i](c)
		}
	}
}

// WithURL returns an OptionFunc to set the Config URL as string `url`
func WithURL(url string) OptionFunc {
	if url == "" {
		return nil
	}
	return func(c *Config) {
		c.URL = url
	}
}

// WithMonitorMode returns an OptionFunc to set the Config Mode as Monitor
func WithMonitorMode(peaks ...int) OptionFunc {
	if len(peaks) == 0 {
		return setMode(Monitor)
	}

	return MultiOption(
		setPeaks(peaks),
		setMode(Monitor),
	)
}

// WithRecordMode returns an OptionFunc to set the Config Mode as Record, configuring it with a string `path` and a
// time.Duration `recTime`
func WithRecordMode(path string, recTime time.Duration) OptionFunc {
	if path == "" {
		return nil
	}
	if recTime == 0 {
		recTime = defaultRecTime
	}
	return MultiOption(
		setMode(Record),
		setDir(path),
		setRecTime(recTime),
	)
}

// WithFilterMode returns an OptionFunc to set the Config Mode as Filter, configuring it with an int `peak`,
// a string `path` and a time.Duration `recTime`
func WithFilterMode(peak []int, path string, recTime time.Duration) OptionFunc {
	if len(peak) == 0 || path == "" {
		return nil
	}
	if recTime == 0 {
		recTime = defaultRecTime
	}
	return MultiOption(
		setPeaks(peak),
		setMode(Filter),
		setDir(path),
		setRecTime(recTime),
	)
}

func WithAnalyzerMode() OptionFunc {
	return setMode(Analyze)
}

// WithMode returns an OptionFunc to set the Config Mode based on string `mode`; accepting also a pointer to
// an int `peak` (if Filter is chosen); accepting a pointer to a string `path` (if Filter or Record is chosen);
// and a pointer to a time.Duration `recTime` (if Filter or Record is chosen)
func WithMode(mode string, peak []int, path *string, recTime *time.Duration) OptionFunc {
	switch mode {
	case "monitor":
		return WithMonitorMode(peak...)
	case "record":
		if path == nil || recTime == nil {
			return nil
		}
		return WithRecordMode(*path, *recTime)
	case "filter":
		if len(peak) == 0 || path == nil || recTime == nil {
			return nil
		}
		return WithFilterMode(peak, *path, *recTime)
	case "analyze":
		return WithAnalyzerMode()
	default:
		return nil
	}
}

// WithDuration returns an OptionFunc to set the Config Dur as time.Duration `dur`
func WithDuration(dur time.Duration) OptionFunc {
	return func(c *Config) {
		c.Dur = &dur
	}
}

// WithRatio returns an OptionFunc to set the Config BufferSize as float64 `ratio`
func WithRatio(ratio float64) OptionFunc {
	if ratio <= 0 {
		return nil
	}
	return func(c *Config) {
		c.BufferSize = ratio
	}
}

// WithPrometheus returns an OptionFunc to set the Config Prom as bool `v`
func WithPrometheus(v bool) OptionFunc {
	return func(c *Config) {
		c.Prom = v
	}
}

// WithPort returns an OptionFunc to set the Config Port (for the Prometheus metrics server) as int `v`
func WithPort(v int) OptionFunc {
	return func(c *Config) {
		c.Port = v
	}
}

// WithExitCode returns an OptionFunc to set the Config ExitCode as int `v`
func WithExitCode(v int) OptionFunc {
	return func(c *Config) {
		c.ExitCode = v
	}
}

func setMode(m Mode) OptionFunc {
	return func(c *Config) {
		c.Mode = m
	}
}

func setPeaks(p []int) OptionFunc {
	return func(c *Config) {
		c.Peak = p
	}
}

func setDir(d string) OptionFunc {
	if string(d[len(d)-4:]) == ".wav" {
		d = d[:len(d)-4]
	}
	return func(c *Config) {
		c.Dir = &d
	}
}

func setRecTime(t time.Duration) OptionFunc {
	return func(c *Config) {
		c.RecTime = &t
	}
}
