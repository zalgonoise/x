package cli

import "github.com/zalgonoise/cfg"

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

	return r
}

func defaultConfig[T FlagType]() Config[T] {
	return Config[T]{
		name: defaultName,
	}
}

func noOpIsValid[T FlagType](*T) error { return nil }

func WithFlag[T FlagType](flag string, zero T, description string) cfg.Option[Config[T]] {
	if flag == "" {
		return cfg.NoOp[Config[T]]{}
	}

	return cfg.Register[Config[T]](func(c Config[T]) Config[T] {
		c.flag = flag
		c.zero = zero
		c.description = description

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

func WithExecutors[T FlagType](executors map[T]Executor) cfg.Option[Config[T]] {
	if len(executors) == 0 {
		return cfg.NoOp[Config[T]]{}
	}

	return cfg.Register[Config[T]](func(c Config[T]) Config[T] {
		c.executors = executors

		return c
	})
}
