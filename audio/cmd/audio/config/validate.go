package config

import (
	"github.com/zalgonoise/x/audio/errs"
)

const (
	confDomain = errs.Domain("audio/config")

	ErrEmpty   = errs.Kind("empty")
	ErrInvalid = errs.Kind("invalid")

	ErrURL        = errs.Entity("URL")
	ErrMode       = errs.Entity("operation mode")
	ErrOutput     = errs.Entity("output")
	ErrOutputPath = errs.Entity("output path")
)

var (
	ErrEmptyURL        = errs.New(confDomain, ErrEmpty, ErrURL)
	ErrInvalidMode     = errs.New(confDomain, ErrInvalid, ErrMode)
	ErrInvalidOutput   = errs.New(confDomain, ErrInvalid, ErrOutput)
	ErrEmptyOutputPath = errs.New(confDomain, ErrEmpty, ErrOutputPath)
)

func Validate(c *Config) error {
	switch c.Mode {
	case Monitor:
	// OK state
	default:
		return ErrInvalidMode
	}

	if c.URL == "" {
		return ErrEmptyURL
	}

	switch c.Output {
	case ToLogger:
	// OK state
	case ToFile, ToPrometheus:
		if c.OutputPath == "" {
			return ErrEmptyOutputPath
		}
	default:
		return ErrInvalidOutput
	}

	return nil
}
