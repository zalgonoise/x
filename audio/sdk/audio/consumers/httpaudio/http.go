package httpaudio

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"io"
	"log/slog"
	"net/http"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/audio/sdk/audio"
)

const (
	headerContentTypeKey   = "Content-Type"
	headerContentTypeValue = "audio/wav"
)

type httpConsumer struct {
	logger  *slog.Logger
	metrics audio.ConsumerMetrics
	tracer  trace.Tracer

	cfg Config

	cancel context.CancelFunc
}

// Consume interacts with the audio source to extract its audio content or stream as an io.Reader.
func (c *httpConsumer) Consume(ctx context.Context) (reader io.Reader, err error) {
	ctx, cancel := context.WithCancel(ctx)

	c.logger.DebugContext(ctx, "starting HTTP connection with audio feed",
		slog.String("target", c.cfg.target),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.target, http.NoBody)
	if err != nil {
		c.logger.ErrorContext(ctx, "error preparing request to connect with HTTP audio feed",
			slog.String("target", c.cfg.target),
			slog.String("error", err.Error()),
		)

		cancel()

		return nil, err
	}

	req.Header.Set(headerContentTypeKey, headerContentTypeValue)

	//nolint:bodyclose // in this implementation, the processor will close the reader once it's done.
	res, err := (&http.Client{
		Timeout: c.cfg.timeout,
	}).Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "error connecting with HTTP audio feed",
			slog.String("target", c.cfg.target),
			slog.String("error", err.Error()),
		)

		cancel()

		return nil, err
	}

	c.cancel = cancel

	c.logger.InfoContext(ctx, "connected to HTTP audio feed")

	return res.Body, nil
}

// Shutdown gracefully shuts down the Consumer.
func (c *httpConsumer) Shutdown(ctx context.Context) error {
	c.logger.InfoContext(ctx, "closing connection to HTTP audio feed")

	c.cancel()

	return nil
}

func New(options []cfg.Option[Config],
	logger *slog.Logger, metrics audio.ConsumerMetrics, tracer trace.Tracer,
) (audio.Consumer, error) {
	config := cfg.Set(DefaultConfig(), options...)

	if err := Validate(config); err != nil {
		return audio.NoOpConsumer(), err
	}

	if logger == nil {
		logger = slog.New(audio.NoOpLogHandler())
	}

	if metrics == nil {
		metrics = audio.NoOpConsumerMetrics{}
	}

	if tracer == nil {
		tracer = noop.NewTracerProvider().Tracer("no-op")
	}

	return &httpConsumer{
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
		cfg:     config,
	}, nil
}
