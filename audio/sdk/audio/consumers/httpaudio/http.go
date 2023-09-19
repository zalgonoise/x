package httpaudio

import (
	"context"
	"io"
	"net/http"

	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/cfg"
)

const (
	headerContentTypeKey   = "Content-Type"
	headerContentTypeValue = "audio/wav"
)

type httpConsumer struct {
	cfg Config

	cancel context.CancelFunc
}

// Consume interacts with the audio source to extract its audio content or stream as an io.Reader.
func (c httpConsumer) Consume(ctx context.Context) (reader io.Reader, err error) {
	ctx, cancel := context.WithCancel(ctx)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.target, http.NoBody)
	if err != nil {
		cancel()

		return nil, err
	}

	req.Header.Set(headerContentTypeKey, headerContentTypeValue)

	res, err := (&http.Client{
		Timeout: c.cfg.timeout,
	}).Do(req)
	if err != nil {
		cancel()

		return nil, err
	}

	c.cancel = cancel

	return res.Body, nil
}

// Shutdown gracefully shuts down the Consumer.
func (c httpConsumer) Shutdown(_ context.Context) error {
	c.cancel()

	return nil
}

func New(options ...cfg.Option[Config]) (audio.Consumer, error) {
	config := cfg.Set[Config](defaultConfig, options...)

	if err := Validate(config); err != nil {
		return audio.NoOpConsumer(), err
	}

	return httpConsumer{
		cfg: config,
	}, nil
}
