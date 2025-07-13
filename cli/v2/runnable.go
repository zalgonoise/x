package cli

import (
	"context"
	"fmt"
	"github.com/zalgonoise/cfg"
	"log/slog"
	"os"
)

type Executor interface {
	Exec(ctx context.Context, logger *slog.Logger, args []string) (int, error)
}

type Runnable struct {
	name string

	isValid   func(*string) error
	executors map[string]Executor
}

func NewRunner(name string, opts ...cfg.Option[Config]) *Runnable {
	config := cfg.Set(defaultConfig(), opts...)

	if name != "" {
		config.name = name
	}

	return apply(&Runnable{}, config)
}

func (r *Runnable) Run(logger *slog.Logger) (int, error) {
	ctx := context.Background()

	// allow mapping an empty subcommand as a top-level executor
	//
	// however, no flags can be passed to that executor
	if len(os.Args) <= 1 {
		if exec, ok := r.executors[""]; ok {
			return exec.Exec(ctx, logger, []string{})
		}
	}

	x := os.Args
	_ = x

	if err := r.isValid(&os.Args[1]); err != nil {
		return 1, fmt.Errorf("%w: %v", ErrInvalidOption, os.Args[1])
	}

	exec, ok := r.executors[os.Args[1]]
	if !ok {
		return 1, fmt.Errorf("invalid option")
	}

	return exec.Exec(ctx, logger, os.Args[2:])
}
