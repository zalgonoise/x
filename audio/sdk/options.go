package sdk

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	otelmetric "go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/zalgonoise/x/audio"
)

// instrumentationName names the tracer and meter the package obtains from the configured providers.
const instrumentationName = "github.com/zalgonoise/x/audio/sdk"

// DefaultBufferDuration is how much audio each Buffer carries unless WithBufferDuration says otherwise.
const DefaultBufferDuration = 100 * time.Millisecond

// Option configures a Pipeline.
type Option func(*config) error

type config struct {
	logger     *slog.Logger
	onError    ErrorHandler
	decode     DecodeFunc
	transforms []Transform
	bufferDur  time.Duration
	tracer     trace.TracerProvider
	meter      otelmetric.MeterProvider
}

func newConfig(opts []Option) (config, error) {
	cfg := config{
		logger:    slog.New(slog.DiscardHandler),
		decode:    decodeAny,
		bufferDur: DefaultBufferDuration,
		tracer:    tracenoop.NewTracerProvider(),
		meter:     metricnoop.NewMeterProvider(),
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		if err := opt(&cfg); err != nil {
			return config{}, err
		}
	}

	if cfg.onError == nil {
		logger := cfg.logger
		cfg.onError = func(err error) { logger.ErrorContext(context.Background(), "audio pipeline", slog.Any("error", err)) }
	}

	return cfg, nil
}

func decodeAny(r io.Reader) (audio.SampleReader, error) {
	sr, _, err := audio.Decode(r)

	return sr, err
}

// WithLogger sets the logger. The default discards everything.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) error {
		if logger == nil {
			return fmt.Errorf("%w: nil logger", ErrInvalidOption)
		}

		c.logger = logger

		return nil
	}
}

// WithErrorHandler sets the handler for errors that do not stop the pipeline, such as a failed Export. The default
// logs them at the error level.
func WithErrorHandler(fn ErrorHandler) Option {
	return func(c *config) error {
		if fn == nil {
			return fmt.Errorf("%w: nil error handler", ErrInvalidOption)
		}

		c.onError = fn

		return nil
	}
}

// WithDecoder sets how the source's bytes are decoded, such as raw PCM with a known format. The default is
// audio.Decode.
func WithDecoder(fn DecodeFunc) Option {
	return func(c *config) error {
		if fn == nil {
			return fmt.Errorf("%w: nil decoder", ErrInvalidOption)
		}

		c.decode = fn

		return nil
	}
}

// WithTransforms appends Transforms, applied in order to every Buffer before it is exported.
func WithTransforms(transforms ...Transform) Option {
	return func(c *config) error {
		for i, t := range transforms {
			if t == nil {
				return fmt.Errorf("%w: nil transform %d", ErrInvalidOption, i)
			}
		}

		c.transforms = append(c.transforms, transforms...)

		return nil
	}
}

// WithTracerProvider sets where the pipeline's spans go: one "audio.pipeline.run" span per Run. The default is a no-op
// provider (CLAUDE.md §7 Q12).
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *config) error {
		if tp == nil {
			return fmt.Errorf("%w: nil tracer provider", ErrInvalidOption)
		}

		c.tracer = tp

		return nil
	}
}

// WithMeterProvider sets where the pipeline's own metrics go: the counters audio.pipeline.buffers,
// audio.pipeline.frames and audio.pipeline.export.errors. The default is a no-op provider. These describe the pipeline,
// not the audio: audio analysis is the job of sdk/metric.
func WithMeterProvider(mp otelmetric.MeterProvider) Option {
	return func(c *config) error {
		if mp == nil {
			return fmt.Errorf("%w: nil meter provider", ErrInvalidOption)
		}

		c.meter = mp

		return nil
	}
}

// WithBufferDuration sets how much audio each Buffer carries (at least one frame). The default is
// DefaultBufferDuration. Shorter buffers lower the latency of live analysis; longer ones lower the per-call overhead.
func WithBufferDuration(d time.Duration) Option {
	return func(c *config) error {
		if d <= 0 {
			return fmt.Errorf("%w: buffer duration %s", ErrInvalidOption, d)
		}

		c.bufferDur = d

		return nil
	}
}
