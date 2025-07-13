package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/zalgonoise/x/errs"
)

var ErrInvalidOption = errs.WithDomain(domainErr, ErrInvalid, ErrOption)

type Runner interface {
	Run(*slog.Logger) (int, error)
}

func Run(runner Runner, logger *slog.Logger) {
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
