package batchreg

import (
	"context"
	"time"

	"github.com/zalgonoise/gbuf"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/unitreg"
	"github.com/zalgonoise/x/cfg"
)

const (
	defaultFlushFrequency = time.Second
	defaultMaxBatchSize   = 256
)

type batchRegistry[T any] struct {
	buffer    *gbuf.RingBuffer[T]
	reg       audio.Registerer[T]
	compactor audio.Compactor[T]

	batchSize int
	errCh     chan error
	cancel    context.CancelFunc
}

func (r batchRegistry[T]) Register(value T) error {
	return r.buffer.WriteItem(value)
}

func (r batchRegistry[T]) Load() <-chan T {
	return r.reg.Load()
}

func (r batchRegistry[T]) Shutdown(ctx context.Context) error {
	_ = r.ForceFlush()

	if closer, ok := r.reg.(interface {
		Shutdown(ctx context.Context) error
	}); ok {
		_ = closer.Shutdown(ctx)
	}

	r.cancel()

	return nil
}

func (r batchRegistry[T]) Err() <-chan error {
	return r.errCh
}

func (r batchRegistry[T]) ForceFlush() error {
	length := r.buffer.Len()

	if length == 0 {
		return nil
	}

	if r.batchSize > 0 && length > r.batchSize {
		length = r.batchSize
	}

	if r.compactor != nil {
		data := make([]T, length)
		if _, err := r.buffer.Read(data); err != nil {
			return err
		}

		v, err := r.compactor(data)
		if err != nil {
			return err
		}

		return r.reg.Register(v)
	}

	item, err := r.buffer.ReadItem()
	if err != nil {
		return err
	}

	return r.reg.Register(item)
}

func (r batchRegistry[T]) run(ctx context.Context, flushFrequency time.Duration) {
	defer close(r.errCh)

	ticker := time.NewTicker(flushFrequency)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if err := r.ForceFlush(); err != nil {
				r.errCh <- err
			}

			return
		case <-ticker.C:
			if err := r.ForceFlush(); err != nil {
				r.errCh <- err

				return
			}
		}
	}
}

func New[T any](options ...cfg.Option[Config[T]]) audio.Registerer[T] {
	config := cfg.New(options...)

	if config.maxBatchSize == 0 {
		config.maxBatchSize = defaultMaxBatchSize
	}

	if config.reg == nil {
		config.reg = unitreg.New[T](config.maxBatchSize)
	}

	if config.flushFrequency == 0 {
		config.flushFrequency = defaultFlushFrequency
	}

	ctx, cancel := context.WithCancel(context.Background())

	batchReg := batchRegistry[T]{
		buffer:    gbuf.NewRingBuffer[T](config.maxBatchSize),
		reg:       config.reg,
		compactor: config.compactor,
		batchSize: config.maxBatchSize,
		errCh:     make(chan error),
		cancel:    cancel,
	}

	go batchReg.run(ctx, config.flushFrequency)

	return batchReg
}
