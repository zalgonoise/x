package stream

import (
	"fmt"
	"time"

	"github.com/zalgonoise/x/ptr"
)

type err string

func (e err) Error() string { return (string)(e) }

const (
	ErrEmptyURL         err = "gsp/conf: empty URL"
	ErrModeNotSupported err = "gsp/conf: operation is not supported"
	ErrModeUnset        err = "gsp/conf: operation is undefined"
	ErrRecTimeUnset     err = "gsp/conf: recording time is undefined"
	ErrDurationUnset    err = "gsp/conf: runtime duration is undefined"
	ErrInvalidRatio     err = "gsp/conf: buffer size ratio cannot be zero or below"
	ErrEmptyDirectory   err = "gsp/conf: recording operation without a target file path"
	ErrEmptyThreshold   err = "gsp/conf: peak threshold is unset"
	ErrShortDuration    err = "gsp/conf: runtime duration is shorter than recording time"
)

const (
	defaultRecTime time.Duration = time.Second * 30
)

var Default = Config{
	Mode:       Monitor,
	Dur:        ptr.To(time.Second * 30),
	BufferSize: 0.5,
}

type Config struct {
	URL        string
	Mode       Mode
	Dur        *time.Duration
	RecTime    *time.Duration
	BufferSize float64
	Peak       *int
	Dir        *string
	Prom       bool
}

func (c *Config) Merge(input *Config) *Config {
	if input.URL != "" {
		c.URL = input.URL
	}
	if input.Mode != Unset {
		c.Mode = input.Mode
	}
	if input.Dur != nil {
		c.Dur = input.Dur
	}
	if input.RecTime != nil {
		c.RecTime = input.RecTime
	}
	if input.BufferSize > 0 {
		c.BufferSize = input.BufferSize
	}
	if input.Peak != nil {
		c.Peak = input.Peak
	}
	if input.Dir != nil {
		c.Dir = input.Dir
	}
	if input.Prom {
		c.Prom = input.Prom
	}
	return c
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return ErrEmptyURL
	}

	switch c.Mode {
	case Monitor:
	case Record:
		if c.RecTime == nil {
			return ErrRecTimeUnset
		}
		if c.Dir == nil {
			return ErrEmptyDirectory
		}
	case Filter:
		if c.RecTime == nil {
			return ErrRecTimeUnset
		}
		if c.Peak == nil || *c.Peak == 0 {
			return ErrEmptyThreshold
		}
		if c.Dir == nil {
			return ErrEmptyDirectory
		}
	case Unset:
		return ErrModeUnset
	default:
		return fmt.Errorf("%w: invalid mode: %d", ErrModeNotSupported, c.Mode)
	}

	if c.Dur == nil {
		return ErrDurationUnset
	}

	if c.RecTime != nil && *c.Dur < *c.RecTime {
		return ErrShortDuration
	}

	if c.BufferSize <= 0 {
		return ErrInvalidRatio
	}

	return nil
}

func newConfig(opts ...Option) *Config {
	conf := &Default

	for _, opt := range opts {
		if opt != nil {
			opt.Apply(conf)
		}
	}

	return conf
}

type Option interface {
	// Apply sets the configuration on the input Config `c`
	Apply(c *Config)
}

func NewConfig(opts ...Option) (*Config, error) {
	c := newConfig(opts...)
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}
