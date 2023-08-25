package httpaudio

import (
	"context"
	"io"

	"github.com/zalgonoise/x/audio/sdk/audio"
)

type httpConsumer struct {
	cfg HTTPConfig
}

// Consume interacts with the audio source to extract its audio content or stream as an io.Reader.
func (c httpConsumer) Consume(ctx context.Context) (reader io.Reader, err error) {
	return nil, nil
}

// Shutdown gracefully shuts down the Consumer.
func (c httpConsumer) Shutdown(ctx context.Context) error {
	return nil
}

func NewConsumer(options ...audio.Option[HTTPConfig]) audio.Consumer {
	return httpConsumer{
		cfg: newHTTPConfig(options...),
	}
}
