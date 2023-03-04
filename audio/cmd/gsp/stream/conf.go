package stream

import (
	"fmt"
	"time"
)

type err string

func (e err) Error() string { return (string)(e) }

const (
	ErrEmptyURL         err = "gsp/conf: empty URL"
	ErrModeNotSupported err = "gsp/conf: operation is not supported"
	ErrEmptyDirectory   err = "gsp/conf: recording operation without a target file path"
	ErrEmptyThreshold   err = "gsp/conf: peak threshold is unset"
	ErrShortDuration    err = "gsp/conf: runtime duration is shorter than recording time"
)

const (
	defaultRecTime time.Duration = time.Second * 30
)

type Config struct {
	URL        string
	Mode       Mode
	Dur        *time.Duration
	RecTime    *time.Duration
	BufferSize float64
	Peak       *int
	Dir        *string
	Term       bool
}

func NewConfig(url, mod string, bufferSize float64, dur, recTime, dir *string, peak *int, term bool) (*Config, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}
	c := new(Config)
	c.URL = url

	if err := c.validateMode(mod, peak); err != nil {
		return nil, err
	}

	c.BufferSize = bufferSize
	if c.BufferSize == 0 {
		c.BufferSize = 1.0
	}

	if err := c.validateDur(dur, recTime); err != nil {
		return nil, err
	}
	if err := c.validateDir(dir); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) validateMode(mod string, peak *int) error {
	switch mod {
	case "monitor":
		c.Mode = Monitor
	case "record":
		c.Mode = Record
	case "filter":
		c.Mode = Filter
		if peak == nil || *peak == 0 {
			return ErrEmptyThreshold
		}
		c.Peak = peak
	default:
		return fmt.Errorf("%w: invalid mode: %s", ErrModeNotSupported, mod)
	}
	return nil
}

func (c *Config) validateDur(dur, recTime *string) error {
	switch c.Mode {
	case Record, Filter:
		var rt = defaultRecTime
		if recTime != nil {
			d, err := time.ParseDuration(*recTime)
			if err != nil {
				return err
			}
			rt = d
		}
		c.RecTime = &rt
	}

	if dur != nil {
		d, err := time.ParseDuration(*dur)
		if err != nil {
			return err
		}
		c.Dur = &d
	}

	if c.Dur != nil && c.RecTime != nil && *c.Dur < *c.RecTime {
		return ErrShortDuration
	}

	return nil
}

func (c *Config) validateDir(dir *string) error {
	if dir == nil || *dir == "" {
		switch c.Mode {
		case Monitor:
			c.Term = true
			return nil
		default:
			return ErrEmptyDirectory
		}
	}
	dirExt := *dir
	if string(dirExt[len(dirExt)-4:]) == ".wav" {
		dirExt = string(dirExt[:len(dirExt)-4])
		dir = &dirExt
	}
	c.Dir = dir
	return nil
}
