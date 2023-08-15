package stream

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/zalgonoise/x/audio/errs"
)

const (
	defaultConnTimeout = 30 * time.Second

	consumerDomain = errs.Domain("audio/stream/consumer")

	ErrEmpty   = errs.Kind("empty")
	ErrAddress = errs.Entity("address")
)

var ErrEmptyAddress = errs.New(consumerDomain, ErrEmpty, ErrAddress)

type HTTPConsumer struct {
	address string
	timeout time.Duration

	cancel context.CancelCauseFunc
}

func (i *HTTPConsumer) Consume(ctx context.Context) (io.Reader, error) {
	ctx, cancel := context.WithCancelCause(ctx)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, i.address, http.NoBody)
	if err != nil {
		cancel(err)

		return nil, err
	}

	req.Header.Set("Content-Type", "audio/wav")

	res, err := (&http.Client{
		Timeout: i.timeout,
	}).Do(req)
	if err != nil {
		cancel(err)

		return nil, err
	}

	i.cancel = cancel

	return res.Body, nil
}

func (i *HTTPConsumer) Shutdown(ctx context.Context) error {
	i.cancel(ctx.Err())

	return nil
}

func NewHTTPConsumer(address string, timeout time.Duration) (*HTTPConsumer, error) {
	if address == "" {
		return nil, ErrEmptyAddress
	}

	if timeout == 0 {
		timeout = defaultConnTimeout
	}

	return &HTTPConsumer{
		address: address,
		timeout: timeout,
	}, nil
}
