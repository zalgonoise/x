package executor

import (
	"context"
	"errors"
	"strings"
	"time"
)

type multiExecutor struct {
	execs []Executor
}

func Multi(execs ...Executor) Executor {
	switch len(execs) {
	case 0:
		return NoOp()
	case 1:
		return execs[0]
	}

	e := make([]Executor, 0, len(execs))

	for i := range execs {
		switch v := execs[i].(type) {
		case nil:
			continue
		case Executable:
			e = append(e, v)
		case multiExecutor:
			e = append(e, v.execs...)
		}
	}

	switch len(e) {
	case 0:
		return NoOp()
	case 1:
		return e[0]
	default:
		return multiExecutor{
			execs: e,
		}
	}
}

func (e multiExecutor) Exec(ctx context.Context) error {
	errs := make([]error, 0, len(e.execs))

	for i := range e.execs {
		if err := e.execs[i].Exec(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (e multiExecutor) Next(ctx context.Context) time.Time {
	return e.execs[0].Next(ctx)
}

func (e multiExecutor) ID() string {
	sb := &strings.Builder{}

	for i := range e.execs {
		if i > 0 {
			sb.WriteByte(':')
		}

		sb.WriteString(e.execs[i].ID())
	}

	return sb.String()
}
