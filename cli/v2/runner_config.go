package cli

import (
	"slices"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/errs"
)

const defaultName = "main"

const (
	domainErr = errs.Domain("x/cli")

	ErrUnsupported = errs.Kind("unsupported")
	ErrInvalid     = errs.Kind("invalid")
	ErrParameter   = errs.Entity("parameter")
	ErrOption      = errs.Entity("option")
)

var ErrUnsupportedParameter = errs.WithDomain(domainErr, ErrUnsupported, ErrParameter)

type Config struct {
	name string

	isValid   func(value *string) error
	executors map[string]Executor
}

func apply(r *Runnable, c Config) *Runnable {
	if c.isValid == nil {
		c.isValid = noOpIsValid
	}

	r.isValid = c.isValid
	r.name = c.name
	r.executors = c.executors

	return r
}

func defaultConfig() Config {
	return Config{
		name: defaultName,
	}
}

func noOpIsValid(*string) error { return nil }

func WithExecutors(executors map[string]Executor) cfg.Option[Config] {
	if len(executors) == 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(c Config) Config {
		c.executors = executors

		selectors := make([]string, 0, len(executors))
		for item := range executors {
			selectors = append(selectors, item)
		}

		c.isValid = func(value *string) error {
			if value == nil || !slices.Contains(selectors, *value) {
				return ErrUnsupportedParameter
			}

			return nil
		}

		return c
	})
}
