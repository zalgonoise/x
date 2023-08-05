package config

import (
	"github.com/zalgonoise/x/audio/errs"
)

const (
	confDomain = errs.Domain("audio/config")

	ErrMissing = errs.Kind("missing")
	ErrEmpty   = errs.Kind("empty")
	ErrInvalid = errs.Kind("invalid")

	ErrConfig     = errs.Entity("configuration")
	ErrURL        = errs.Entity("URL")
	ErrMode       = errs.Entity("operation mode")
	ErrOutput     = errs.Entity("output")
	ErrOutputPath = errs.Entity("output path")
	ErrNumBuckets = errs.Entity("number of spectrum buckets")
)

var (
	ErrMissingConfig   = errs.New(confDomain, ErrMissing, ErrConfig)
	ErrEmptyURL        = errs.New(confDomain, ErrEmpty, ErrURL)
	ErrInvalidMode     = errs.New(confDomain, ErrInvalid, ErrMode)
	ErrInvalidOutput   = errs.New(confDomain, ErrInvalid, ErrOutput)
	ErrEmptyOutputPath = errs.New(confDomain, ErrEmpty, ErrOutputPath)
	ErrEmptyNumBuckets = errs.New(confDomain, ErrEmpty, ErrNumBuckets)
)

const minNumBuckets = 8

// Validate returns an error if the input Config contains invalid data
func Validate(c *Config) error {
	switch c.Mode {
	case Monitor:
	// OK state
	case Analyze, Combined:
		if c.NumSpectrumBuckets < minNumBuckets {
			return ErrEmptyNumBuckets
		}
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
