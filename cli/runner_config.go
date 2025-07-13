package cli

import (
	"slices"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/errs"
)

const (
	domainErr = errs.Domain("x/cli")

	ErrUnsupported = errs.Kind("unsupported")

	ErrParameter = errs.Entity("parameter")
)

var ErrUnsupportedParameter = errs.WithDomain(domainErr, ErrUnsupported, ErrParameter)

type Config[T FlagType] struct {
	name string

	flag        string
	zero        T
	description string

	isValid   func(*T) error
	executors map[T]Executor
}

func apply[T FlagType](r Runner, c Config[T]) Runner {
	run, ok := r.(runner[T])
	if !ok {
		return NoOp()
	}

	if c.isValid == nil {
		c.isValid = noOpIsValid[T]
	}

	run.name = c.name
	run.flag = c.flag
	run.zero = c.zero
	run.description = c.description
	run.isValid = c.isValid
	run.executors = c.executors

	return run
}

func defaultConfig[T FlagType]() Config[T] {
	return Config[T]{
		name: defaultName,
	}
}

func noOpIsValid[T FlagType](*T) error { return nil }

func WithFlag[T FlagType](flag string, zero T, description string, executor Executor) cfg.Option[Config[T]] {
	if flag == "" {
		return cfg.NoOp[Config[T]]{}
	}

	return cfg.Register[Config[T]](func(c Config[T]) Config[T] {
		c.flag = flag
		c.zero = zero
		c.description = description

		if c.executors == nil {
			c.executors = make(map[T]Executor)
		}

		c.executors[zero] = executor

		return c
	})
}

func WithValidation[T FlagType](isValid func(*T) error) cfg.Option[Config[T]] {
	if isValid == nil {
		return cfg.NoOp[Config[T]]{}
	}

	return cfg.Register[Config[T]](func(c Config[T]) Config[T] {
		c.isValid = isValid

		return c
	})
}

func WithOneOf[T FlagType](selectors ...T) cfg.Option[Config[T]] {
	if len(selectors) == 0 {
		return cfg.NoOp[Config[T]]{}
	}

	return cfg.Register[Config[T]](func(c Config[T]) Config[T] {
		c.isValid = func(value *T) error {
			if value == nil || !slices.Contains(selectors, *value) {
				return ErrUnsupportedParameter
			}

			return nil
		}

		return c
	})
}

func WithExecutors[T FlagType](executors map[T]Executor) cfg.Option[Config[T]] {
	if len(executors) == 0 {
		return cfg.NoOp[Config[T]]{}
	}

	return cfg.Register[Config[T]](func(c Config[T]) Config[T] {
		c.executors = executors

		selectors := make([]T, 0, len(executors))
		for item := range executors {
			selectors = append(selectors, item)
		}

		c.isValid = func(value *T) error {
			if value == nil || !slices.Contains(selectors, *value) {
				return ErrUnsupportedParameter
			}

			return nil
		}

		return c
	})
}
