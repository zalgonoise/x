package stream

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/zalgonoise/x/audio/fft"
)

const defaultFlushFrequency = time.Second

type AggregatorConfig struct {
	FlushFrequency time.Duration
	MaxBatchSize   int
}

type aggregation[T cmp.Ordered] struct {
	mu    *sync.Mutex
	n     int
	value T
}

type dynAggregation[T any] struct {
	mu       *sync.Mutex
	n        int
	value    T
	lessFunc func(i T, j T) bool
}

type MaxAggregator struct {
	peaks   *aggregation[float64]
	spectra *dynAggregation[fft.FrequencyPower]

	exporter Exporter
	config   *AggregatorConfig
}

func (a MaxAggregator) SendPeak(value float64) (err error) {
	if value == 0.0 {
		return
	}

	a.peaks.mu.Lock()
	a.peaks.n++
	if a.peaks.value < value {
		a.peaks.value = value
	}
	a.peaks.mu.Unlock()

	if a.peaks.n >= a.config.MaxBatchSize {
		if err = a.ForceFlush(context.Background()); err != nil {
			return err
		}
	}

	return nil
}

func (a MaxAggregator) SendSpectrum(frequencies []fft.FrequencyPower) (err error) {
	slices.SortFunc(frequencies, func(a, b fft.FrequencyPower) int {
		switch {
		case a.Mag < b.Mag:
			return -1
		case a.Mag > b.Mag:
			return 1
		default:
			return 0
		}
	})

	a.spectra.mu.Lock()
	a.spectra.n++
	if a.spectra.value.Mag < frequencies[0].Mag {
		a.spectra.value = frequencies[0]
	}
	a.spectra.mu.Unlock()

	if a.spectra.n >= a.config.MaxBatchSize {
		if err = a.ForceFlush(context.Background()); err != nil {
			return err
		}
	}

	return nil
}

func (a MaxAggregator) ForceFlush(_ context.Context) error {
	errs := make([]error, 0, 2)

	if err := a.exporter.SendPeak(a.peaks.value); err != nil {
		errs = append(errs, err)
	}

	if err := a.exporter.SendSpectrum([]fft.FrequencyPower{a.spectra.value}); err != nil {
		errs = append(errs, err)
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

func (a MaxAggregator) Shutdown(ctx context.Context) error {
	return a.exporter.Shutdown(ctx)
}

func NewMaxAggregator(
	ctx context.Context, cancel context.CancelCauseFunc,
	exporter Exporter, config *AggregatorConfig,
) *MaxAggregator {
	if config == nil {
		config = new(AggregatorConfig)
	}

	aggregator := &MaxAggregator{
		peaks: &aggregation[float64]{
			mu: new(sync.Mutex),
		},
		spectra: &dynAggregation[fft.FrequencyPower]{
			mu: new(sync.Mutex),
			lessFunc: func(i, j fft.FrequencyPower) bool {
				return i.Mag < j.Mag
			},
		},

		exporter: exporter,
		config:   config,
	}

	if config.FlushFrequency == 0 {
		config.FlushFrequency = defaultFlushFrequency
	}

	go func() {
		ticker := time.NewTicker(config.FlushFrequency)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := aggregator.ForceFlush(ctx); err != nil {
					cancel(err)

					return
				}
			}
		}
	}()

	return aggregator
}
