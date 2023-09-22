package batchreg

import (
	"context"
	"errors"
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
	reg       audio.Registry[T]
	compactor audio.Compactor[T]

	batchSize int
	errCh     chan error
	cancel    context.CancelFunc
}

// Register stores the input data in the audio.Registry's inner buffer, returning an error if raised.
func (r batchRegistry[T]) Register(value T) error {
	return r.buffer.WriteItem(value)
}

// Load returns a receive-only channel of items of a given type, usually as part of a Registry features.
//
// The returned channel is actually the underlying audio.Registry's values channel.
func (r batchRegistry[T]) Load() <-chan T {
	return r.reg.Load()
}

// Shutdown gracefully stops the audio.Registry.
//
// It will both ForceFlush and call its inner audio.Registry's Shutdown method, returning any error if raised.
func (r batchRegistry[T]) Shutdown(ctx context.Context) error {
	defer r.cancel()
	errs := make([]error, 0, 2)

	if err := r.ForceFlush(); err != nil {
		errs = append(errs, err)
	}

	if err := r.reg.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// Err provides access to this audio.Registry's error channel, to provide
// visibility over its runtime errors.
func (r batchRegistry[T]) Err() <-chan error {
	return r.errCh
}

func (r batchRegistry[T]) flushToCompactor() error {
	length := r.buffer.Len()
	if length == 0 {
		return nil
	}

	if r.batchSize > 0 && length > r.batchSize {
		length = r.batchSize
	}

	data := make([]T, length)
	if _, err := r.buffer.Read(data); err != nil {
		return err
	}

	v, err := r.compactor(data)
	if err != nil {
		return err
	}

	if err = r.reg.Register(v); err != nil {
		return err
	}

	if r.buffer.Len() > 0 {
		return r.flushToCompactor()
	}

	return nil
}

// ForceFlush checks if the audio.Registry's value buffer contains any items, flushing them
// to the underlying audio.Registry if applicable.
//
// If an audio.Compactor is configured, the existing items are reduced with it in batches, if configured.
// The audio.Registry goes through multiple passes through the data if necessary, while there are items in the buffer.
//
// Otherwise, the latest value is registered, instead, and the buffer is drained.
func (r batchRegistry[T]) ForceFlush() error {
	if r.compactor != nil {
		return r.flushToCompactor()
	}

	length := r.buffer.Len()

	if length > 0 {
		items := make([]T, length)

		if _, err := r.buffer.Read(items); err != nil {
			return err
		}

		return r.reg.Register(items[len(items)-1])
	}

	return nil
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

func New[T any](options ...cfg.Option[Config[T]]) audio.Registry[T] {
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
