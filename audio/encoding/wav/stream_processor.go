package wav

import (
	"context"
	"errors"
)

// ProcessFunc describes a function that processes a portion of the audio buffer
// as it is read and decoded from the incoming byte stream.
type ProcessFunc func(header *Header, data []float64) error

// ProcessContextFunc describes a function that processes a portion of the audio buffer
// as it is read and decoded from the incoming byte stream, and also accepts a context.
type ProcessContextFunc func(ctx context.Context, header *Header, data []float64) error

func processFuncWithContext(fn ProcessFunc) ProcessContextFunc {
	return func(_ context.Context, header *Header, data []float64) error {
		return fn(header, data)
	}
}

// MultiProc merges multiple processor functions for floating point audio data, with
// or without a fail-fast strategy.
func MultiProc(failFast bool, fns ...ProcessFunc) ProcessFunc {
	switch len(fns) {
	case 0:
		return nil
	case 1:
		return fns[0]
	}

	if failFast {
		return func(h *Header, data []float64) error {
			for i := range fns {
				if err := fns[i](h, data); err != nil {
					return err
				}
			}

			return nil
		}
	}

	return func(h *Header, data []float64) error {
		errs := make([]error, 0, len(fns))

		for i := range fns {
			if err := fns[i](h, data); err != nil {
				errs = append(errs, err)
			}
		}

		switch len(errs) {
		case 0:
			return nil
		case 1:
			return errs[0]
		default:
			return errors.Join(errs...)
		}
	}
}

func ErrorPipe(fn ProcessFunc, errs chan<- error) ProcessFunc {
	return func(header *Header, data []float64) error {
		err := fn(header, data)
		if err != nil {
			errs <- err
		}

		return err
	}
}

// MultiProcContext merges multiple processor functions for floating point audio data, with
// or without a fail-fast strategy.
func MultiProcContext(failFast bool, fns ...ProcessContextFunc) ProcessContextFunc {
	switch len(fns) {
	case 0:
		return nil
	case 1:
		return fns[0]
	}

	if failFast {
		return func(ctx context.Context, h *Header, data []float64) error {
			for i := range fns {
				if err := fns[i](ctx, h, data); err != nil {
					return err
				}
			}

			return nil
		}
	}

	return func(ctx context.Context, h *Header, data []float64) error {
		errs := make([]error, 0, len(fns))

		for i := range fns {
			if err := fns[i](ctx, h, data); err != nil {
				errs = append(errs, err)
			}
		}

		switch len(errs) {
		case 0:
			return nil
		case 1:
			return errs[0]
		default:
			return errors.Join(errs...)
		}
	}
}

func ErrorPipeContext(fn ProcessContextFunc, errs chan<- error) ProcessContextFunc {
	return func(ctx context.Context, header *Header, data []float64) error {
		err := fn(ctx, header, data)
		if err != nil {
			errs <- err
		}

		return err
	}
}
