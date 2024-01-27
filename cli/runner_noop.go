package cli

import "log/slog"

type noOpRunner struct{}

func (r noOpRunner) Run(*slog.Logger) (int, error) {
	return 0, nil
}

func NoOp() Runner {
	return noOpRunner{}
}
