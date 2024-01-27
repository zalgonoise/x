package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/zalgonoise/cfg"
)

const defaultName = "main"

var (
	ErrInvalidOption     = errors.New("invalid option")
	ErrUnsupportedOption = errors.New("unsupported option")
)

func Run(runner Runner) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if runner == nil {
		logger.ErrorContext(context.Background(), "nil runner; exiting")
		os.Exit(1)
	}

	code, err := runner.Run(logger)
	if err != nil {
		logger.ErrorContext(context.Background(), "runtime error", slog.String("error", err.Error()))
		os.Exit(code)
	}

	os.Exit(code)
}

type Runner interface {
	Run(*slog.Logger) (int, error)
}

type FlagType interface {
	bool | string | int | int64 | uint | uint64 | float64 | time.Duration
}

type Valuer[T FlagType] func(*flag.FlagSet, string, T, string) *T

func NewRunner[T FlagType](name string, opts ...cfg.Option[Config[T]]) Runner {
	config := cfg.Set(defaultConfig[T](), opts...)

	if name != "" {
		config.name = name
	}

	var r Runner

	switch any(config.zero).(type) {
	case bool:
		r = newBoolRunner()
	case string:
		r = newStringRunner()
	case int:
		r = newIntRunner()
	case int64:
		r = newInt64Runner()
	case uint:
		r = newUintRunner()
	case uint64:
		r = newUint64Runner()
	case float64:
		r = newFloat64Runner()
	case time.Duration:
		r = newDurationRunner()
	default:
	}

	return apply[T](r, config)
}

type runner[T FlagType] struct {
	name string

	flag        string
	zero        T
	description string

	valuer Valuer[T]

	isValid   func(*T) error
	executors map[T]Executor
}

func (r runner[T]) Run(logger *slog.Logger) (int, error) {
	fs := flag.NewFlagSet(r.name, flag.ExitOnError)

	value := r.valuer(fs, r.flag, r.zero, r.description)

	if err := fs.Parse(os.Args[1:3]); err != nil {
		return 1, err
	}

	if err := r.isValid(value); err != nil {
		return 1, fmt.Errorf("%w: %v", ErrInvalidOption, value)
	}

	ctx := context.Background()

	exec, ok := r.executors[*value]
	if !ok {
		return 1, fmt.Errorf("%w: %v", ErrUnsupportedOption, *value)
	}

	return exec.Exec(ctx, logger, os.Args[3:])
}

func newStringRunner() Runner {
	return runner[string]{
		valuer: func(fs *flag.FlagSet, name, zero, description string) *string {
			return fs.String(name, zero, description)
		},
	}
}

func newBoolRunner() Runner {
	return runner[bool]{
		valuer: func(fs *flag.FlagSet, name string, zero bool, description string) *bool {
			return fs.Bool(name, zero, description)
		},
	}
}

func newIntRunner() Runner {
	return runner[int]{
		valuer: func(fs *flag.FlagSet, name string, zero int, description string) *int {
			return fs.Int(name, zero, description)
		},
	}
}

func newInt64Runner() Runner {
	return runner[int64]{
		valuer: func(fs *flag.FlagSet, name string, zero int64, description string) *int64 {
			return fs.Int64(name, zero, description)
		},
	}
}

func newUintRunner() Runner {
	return runner[uint]{
		valuer: func(fs *flag.FlagSet, name string, zero uint, description string) *uint {
			return fs.Uint(name, zero, description)
		},
	}
}

func newUint64Runner() Runner {
	return runner[uint64]{
		valuer: func(fs *flag.FlagSet, name string, zero uint64, description string) *uint64 {
			return fs.Uint64(name, zero, description)
		},
	}
}

func newFloat64Runner() Runner {
	return runner[float64]{
		valuer: func(fs *flag.FlagSet, name string, zero float64, description string) *float64 {
			return fs.Float64(name, zero, description)
		},
	}
}

func newDurationRunner() Runner {
	return runner[time.Duration]{
		valuer: func(fs *flag.FlagSet, name string, zero time.Duration, description string) *time.Duration {
			return fs.Duration(name, zero, description)
		},
	}
}
