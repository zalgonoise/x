package stream

import (
	"time"

	"github.com/zalgonoise/x/ptr"
)

type multiOption struct {
	opts []Option
}

// Apply sets the configuration on the input Config `c`
func (mo *multiOption) Apply(c *Config) {
	for _, opt := range mo.opts {
		opt.Apply(c)
	}
}

// MultiOption merges a slice of Option into one
func MultiOption(opts ...Option) Option {
	if len(opts) == 0 {
		return nil
	}
	mo := new(multiOption)
	for i := range opts {
		if opts[i] == nil {
			continue
		}
		if mopt, ok := opts[i].(*multiOption); ok {
			mo.opts = append(mo.opts, mopt.opts...)
			continue
		}
		mo.opts = append(mo.opts, opts[i])
	}
	return mo
}

type optURL string

// Apply sets the configuration on the input Config `c`
func (o optURL) Apply(c *Config) {
	c.URL = (string)(o)
}

// WithURL returns an Option to set the Config URL as string `url`
func WithURL(url string) Option {
	if url == "" {
		return nil
	}
	return optURL(url)
}

type optMode Mode

// Apply sets the configuration on the input Config `c`
func (o optMode) Apply(c *Config) {
	c.Mode = (Mode)(o)
}

// WithMonitorMode returns an Option to set the Config Mode as Monitor
func WithMonitorMode(peaks ...int) Option {
	if len(peaks) == 0 {
		return optMode(Monitor)
	}

	return MultiOption(
		optPeak(peaks),
		optMode(Monitor),
	)
}

// WithRecordMode returns an Option to set the Config Mode as Record, configuring it with a string `path` and a
// time.Duration `recTime`
func WithRecordMode(path string, recTime time.Duration) Option {
	if path == "" {
		return nil
	}
	if recTime == 0 {
		recTime = defaultRecTime
	}
	return MultiOption(
		optMode(Record),
		optDir(path),
		optRecTime(recTime),
	)
}

// WithFilterMode returns an Option to set the Config Mode as Filter, configuring it with an int `peak`,
// a string `path` and a time.Duration `recTime`
func WithFilterMode(peak []int, path string, recTime time.Duration) Option {
	if len(peak) == 0 || path == "" {
		return nil
	}
	if recTime == 0 {
		recTime = defaultRecTime
	}
	return MultiOption(
		optPeak(peak),
		optMode(Filter),
		optDir(path),
		optRecTime(recTime),
	)
}

// WithMode returns an Option to set the Config Mode based on string `mode`; accepting also a pointer to
// an int `peak` (if Filter is chosen); accepting a pointer to a string `path` (if Filter or Record is chosen);
// and a pointer to a time.Duration `recTime` (if Filter or Record is chosen)
func WithMode(mode string, peak []int, path *string, recTime *time.Duration) Option {
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
	default:
		return nil
	}
}

type optDur time.Duration

// Apply sets the configuration on the input Config `c`
func (o optDur) Apply(c *Config) {
	c.Dur = ptr.To((time.Duration)(o))
}

// WithDuration returns an Option to set the Config Dur as time.Duration `dur`
func WithDuration(dur time.Duration) Option {
	return optDur(dur)
}

type optRecTime time.Duration

// Apply sets the configuration on the input Config `c`
func (o optRecTime) Apply(c *Config) {
	c.RecTime = ptr.To((time.Duration)(o))
}

type optBufferSize float64

// Apply sets the configuration on the input Config `c`
func (o optBufferSize) Apply(c *Config) {
	c.BufferSize = (float64)(o)
}

// WithRatio returns an Option to set the Config BufferSize as float64 `ratio`
func WithRatio(ratio float64) Option {
	if ratio <= 0 {
		return nil
	}
	return optBufferSize(ratio)
}

type optPeak []int

// Apply sets the configuration on the input Config `c`
func (o optPeak) Apply(c *Config) {
	c.Peak = o
}

type optDir string

// Apply sets the configuration on the input Config `c`
func (o optDir) Apply(c *Config) {
	if string(o[len(o)-4:]) == ".wav" {
		o = optDir(o[:len(o)-4])
	}
	c.Dir = ptr.To((string)(o))
}

type optProm bool

// Apply sets the configuration on the input Config `c`
func (o optProm) Apply(c *Config) {
	c.Prom = (bool)(o)
}

// WithPrometheus returns an Option to set the Config Prom as bool `v`
func WithPrometheus(v bool) Option {
	return optProm(v)
}

type optPromPort int

// Apply sets the configuration on the input Config `c`
func (o optPromPort) Apply(c *Config) {
	c.Port = (int)(o)
}

func WithPort(v int) Option {
	return optPromPort(v)
}

type optExitCode int

// Apply sets the configuration on the input Config `c`
func (o optExitCode) Apply(c *Config) {
	c.ExitCode = (int)(o)
}

func WithExitCode(v int) Option {
	return optExitCode(v)
}
