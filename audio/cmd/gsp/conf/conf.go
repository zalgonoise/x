package conf

import (
	"fmt"
	"time"

	"github.com/zalgonoise/x/audio/cmd/gsp/mode"
)

type err string

func (e err) Error() string { return (string)(e) }

const (
	ErrEmptyURL         err = "gsp/conf: empty URL"
	ErrModeNotSupported err = "gsp/conf: operation is not supported"
	ErrEmptyDirectory   err = "gsp/conf: recording operation without a target file path"
	ErrEmptyThreshold   err = "gsp/conf: peak threshold is unset"
)

type Config struct {
	URL  string
	Mode mode.Mode
	Dur  *time.Duration
	Peak *int
	Dir  *string
	Term bool
}

func New(url, mod string, dur, dir *string, peak *int, term bool) (*Config, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}
	c := new(Config)
	c.URL = url

	switch mod {
	case "monitor":
		c.Mode = mode.Monitor
	case "record":
		c.Mode = mode.Record
	case "filter":
		c.Mode = mode.Filter
		if peak == nil || *peak == 0 {
			return nil, ErrEmptyThreshold
		}
		c.Peak = peak
	default:
		return nil, fmt.Errorf("%w: invalid mode: %s", ErrModeNotSupported, mod)
	}

	if dur != nil {
		d, err := time.ParseDuration(*dur)
		if err != nil {
			return nil, err
		}
		c.Dur = &d
	}

	if dir == nil || *dir == "" {
		switch c.Mode {
		case mode.Monitor:
			c.Term = true
		default:
			return nil, ErrEmptyDirectory
		}
	}
	c.Dir = dir
	return c, nil
}
