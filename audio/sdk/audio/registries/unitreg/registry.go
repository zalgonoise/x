package unitreg

import (
	"context"

	"github.com/zalgonoise/x/audio/sdk/audio"
)

const defaultBufferSize = 64

type unitRegistry[T any] struct {
	ch chan T
}

func (r *unitRegistry[T]) Register(value T) error {
	r.ch <- value

	return nil
}

func (r *unitRegistry[T]) Load() <-chan T {
	return r.ch
}

func (r *unitRegistry[T]) Shutdown(context.Context) error {
	close(r.ch)

	return nil
}

func New[T any](size int) audio.Registry[T] {
	if size < 0 {
		size = defaultBufferSize
	}

	return &unitRegistry[T]{
		ch: make(chan T, size),
	}
}
