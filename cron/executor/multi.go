package executor

import (
	"context"
	"errors"
	"sync"
)

func Multi(ctx context.Context, execs ...Executor) error {
	errs := make([]error, 0, len(execs))
	wg := &sync.WaitGroup{}

	for i := range execs {
		i := i

		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := execs[i].Exec(ctx); err != nil {
				errs = append(errs, err)
			}
		}()
	}

	wg.Wait()

	return errors.Join(errs...)
}
