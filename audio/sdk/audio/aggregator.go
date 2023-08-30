package audio

import (
	"context"
	"io"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zalgonoise/gbuf"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/wav/header"
	"github.com/zalgonoise/x/cfg"
)

// Collector is a generic type that is able to parse incoming audio chunks to retrieve
// meaningful information about the signal.
//
// A Collector can process the audio data and extract whatever it wants, creating suitable
// metrics for its needs.
//
// It is the responsibility of the Exporter to store collected values and push them to the
// appropriate backend
type Collector interface {
	Collect(h *header.Header, data []float64) error
}

// Registerer is a generic type that registers and loads values on a specific type context.
//
// Registries are responsible for handling aggregations and compacting values into one, when Load is called
type Registerer[T any] interface {
	Register(T) error
	Load() (T, error)
}

type collector[T any] struct {
	extractor  Extractor[T]
	registerer Registerer[T]
}

// Collect implements the Collector interface.
//
// It will use its inner Registerer and Extractor to register the extracted value from the input.
func (a collector[T]) Collect(h *header.Header, data []float64) error {
	return a.registerer.Register(a.extractor.Extract(h, data))
}

// NewCollector creates a Collector from hte input Extractor and Registerer
func NewCollector[T any](extractor Extractor[T], registerer Registerer[T]) Collector {
	return collector[T]{
		extractor:  extractor,
		registerer: registerer,
	}
}

// Compactor is a function that creates a summary of a set of values based on a certain rule (max, average, rate, etc)
// returning one single value of the same type and an error if raised.
type Compactor[T any] func([]T) (T, error)

// Extraction is a generic function type that serves as an audio processor function,
// but returns any type desired, as appropriate to the analysis, processing, recording, whatever it may be.
//
// It is of the responsibility of the Exporter to position the configured Extractor to actually export the
// aggregations.
//
// The sole responsibility of an Extractor is to convert raw audio (as chunks of float64 values) into anything
// meaningful, that is exported / handled separately. Not all Exporter will need one or more Extractor, however
// these are supposed to be perceived as preset building blocks to work with the incoming audio chunks.
type Extraction[T any] func(*header.Header, []float64) T

func (a Extraction[T]) Extract(h *header.Header, data []float64) T {
	return a(h, data)
}

// Extractor is a generic interface for a type that implements the Extract method, which can return a value from
// parsing an audio chunk.
type Extractor[T any] interface {
	Extract(*header.Header, []float64) T
}

// MaxPeak returns a float64 Collector that calculates the maximum peak value in an audio signal
func MaxPeak() Extractor[float64] {
	return Extraction[float64](func(_ *header.Header, data []float64) (maximum float64) {
		for i := range data {
			if data[i] > maximum {
				maximum = data[i]
			}
		}

		return maximum
	})
}

// AveragePeak returns a float64 Collector that calculates the average peak value in an audio signal
func AveragePeak() Extractor[float64] {
	return Extraction[float64](func(_ *header.Header, data []float64) (average float64) {
		for i := range data {
			average += data[i]
		}

		return average / float64(len(data))
	})
}

// MaxSpectrum returns a []fft.FrequencyPower Collector that calculates the maximum spectrum values in an audio signal
func MaxSpectrum(size int) Extractor[[]fft.FrequencyPower] {
	if size < 8 {
		size = 64
	}

	sampleRate := 44100

	return Extraction[[]fft.FrequencyPower](func(h *header.Header, data []float64) []fft.FrequencyPower {
		if h != nil {
			sampleRate = int(h.SampleRate)
		}

		bs := fft.NearestBlock(size)
		windowBlock := window.New(window.Blackman, int(bs))

		maximum := make([]fft.FrequencyPower, 0, len(data)/int(bs))

		for i := 0; i+int(bs) < len(data); i += int(bs) {
			spectrum := fft.Apply(
				sampleRate,
				data[i:i+int(bs)],
				windowBlock,
			)

			slices.SortFunc(spectrum, func(a, b fft.FrequencyPower) int {
				switch {
				case a.Mag > b.Mag:
					return -1
				case a.Mag < b.Mag:
					return 1
				default:
					return 0
				}
			})

			maximum = append(maximum, spectrum[0])
		}

		return maximum
	})
}

type BatchConfig struct {
	flushFrequency time.Duration
	maxBatchSize   int
}

func WithBatchSize(size int) cfg.Option[BatchConfig] {
	return cfg.Register(func(config BatchConfig) BatchConfig {
		config.maxBatchSize = size

		return config
	})
}

func WithFlushFrequency(dur time.Duration) cfg.Option[BatchConfig] {
	return cfg.Register(func(config BatchConfig) BatchConfig {
		config.flushFrequency = dur

		return config
	})
}

type unitRegistry[T any] struct {
	value T
	mu    *sync.Mutex
	isSet *atomic.Bool
}

func (r *unitRegistry[T]) Register(value T) error {
	r.mu.Lock()
	r.value = value
	r.isSet.Store(true)
	r.mu.Unlock()

	return nil
}

func (r *unitRegistry[T]) Load() (T, error) {
	r.mu.Lock()
	value := r.value
	isSet := r.isSet.Load()
	r.isSet.Store(false)
	r.mu.Unlock()

	if !isSet {
		return *new(T), io.EOF
	}

	return value, nil
}

func NewRegistry[T any]() Registerer[T] {
	return &unitRegistry[T]{
		mu:    &sync.Mutex{},
		isSet: &atomic.Bool{},
	}
}

type batchRegistry[T any] struct {
	buffer    *gbuf.RingBuffer[T]
	reg       Registerer[T]
	compactor Compactor[T]

	config BatchConfig

	errCh  chan error
	cancel context.CancelFunc
}

func (r batchRegistry[T]) Register(value T) error {
	return r.buffer.WriteItem(value)
}

func (r batchRegistry[T]) Load() (T, error) {
	if r.compactor != nil {
		return r.compactor(r.buffer.Value())
	}

	return r.buffer.ReadItem()
}

func (r batchRegistry[T]) Shutdown(_ context.Context) error {
	r.cancel()

	return nil
}

func (r batchRegistry[T]) Err() <-chan error {
	return r.errCh
}

func (r batchRegistry[T]) flush() error {
	value := r.buffer.Value()

	if r.compactor != nil {
		v, err := r.compactor(value)
		if err != nil {
			return err
		}

		return r.reg.Register(v)
	}

	for i := range value {
		if err := r.reg.Register(value[i]); err != nil {
			return err
		}
	}

	return nil
}

func BatchRegistry[T any](reg Registerer[T], compactor Compactor[T], options ...cfg.Option[BatchConfig]) Registerer[T] {
	config := cfg.New[BatchConfig[T]](options...)

	ctx, cancel := context.WithCancel(context.Background())

	batchReg := batchRegistry[T]{
		buffer:    gbuf.NewRingBuffer[T](config.maxBatchSize),
		reg:       reg,
		config:    config,
		compactor: compactor,
		errCh:     make(chan error),
		cancel:    cancel,
	}

	go func() {
		ticker := time.NewTicker(config.flushFrequency)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				if err := batchReg.flush(); err != nil {
					batchReg.errCh <- err
				}

				return
			case <-ticker.C:
				if err := batchReg.flush(); err != nil {
					batchReg.errCh <- err

					return
				}
			default:
				if config.maxBatchSize > 0 && batchReg.buffer.Len() > config.maxBatchSize {
					if err := batchReg.flush(); err != nil {
						batchReg.errCh <- err

						return
					}
				}
			}
		}
	}()

	return batchReg
}
