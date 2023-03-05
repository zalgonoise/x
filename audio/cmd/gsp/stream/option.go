package stream

import (
	"time"

	"github.com/zalgonoise/x/ptr"
)

type multiOption struct {
	opts []Option
}

func (mo *multiOption) Apply(c *Config) {
	for _, opt := range mo.opts {
		opt.Apply(c)
	}
}

func newMultiOption(opts ...Option) Option {
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

func (o optURL) Apply(c *Config) {
	c.URL = (string)(o)
}

func WithURL(url string) Option {
	return optURL(url)
}

type optMode Mode

func (o optMode) Apply(c *Config) {
	c.Mode = (Mode)(o)
}

func WithMonitorMode() Option {
	return optMode(Monitor)
}
func WithRecordMode(path string, recTime time.Duration) Option {
	if path == "" || recTime == 0 {
		return nil
	}
	return newMultiOption(
		optMode(Record),
		optDir(path),
		optRecTime(recTime),
	)
}
func WithFilterMode(peak int, path string, recTime time.Duration) Option {
	if peak == 0 || path == "" || recTime == 0 {
		return nil
	}
	return newMultiOption(
		optPeak(peak),
		optMode(Filter),
		optDir(path),
		optRecTime(recTime),
	)
}

func WithMode(mode string, peak *int, path *string, recTime *time.Duration) Option {
	switch mode {
	case "monitor":
		return WithMonitorMode()
	case "record":
		if path == nil || recTime == nil {
			return nil
		}
		return WithRecordMode(*path, *recTime)
	case "filter":
		if peak == nil || path == nil || recTime == nil {
			return nil
		}
		return WithFilterMode(*peak, *path, *recTime)
	default:
		return nil
	}
}

type optDur time.Duration

func (o optDur) Apply(c *Config) {
	c.Dur = ptr.To((time.Duration)(o))
}

func WithDuration(dur time.Duration) Option {
	return optDur(dur)
}

type optRecTime time.Duration

func (o optRecTime) Apply(c *Config) {
	c.RecTime = ptr.To((time.Duration)(o))
}

type optBufferSize float64

func (o optBufferSize) Apply(c *Config) {
	c.BufferSize = (float64)(o)
}

func WithRatio(ratio float64) Option {
	if ratio <= 0 {
		return nil
	}
	return optBufferSize(ratio)
}

type optPeak int

func (o optPeak) Apply(c *Config) {
	c.Peak = ptr.To((int)(o))
}

type optDir string

func (o optDir) Apply(c *Config) {
	if string(o[len(o)-4:]) == ".wav" {
		o = optDir(o[:len(o)-4])
	}
	c.Dir = ptr.To((string)(o))
}

type optProm bool

func (o optProm) Apply(c *Config) {
	c.Prom = (bool)(o)
}
func WithPrometheus(v bool) Option {
	return optProm(v)
}
