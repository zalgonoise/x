package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"

	"github.com/zalgonoise/x/audio/errs"
)

const (
	domain = errs.Domain("audio/http")

	ErrEmpty     = errs.Kind("empty")
	ErrExhausted = errs.Kind("exhausted")

	ErrURL     = errs.Entity("URL")
	ErrBackoff = errs.Entity("backoff")
)

var (
	ErrEmptyURL         = errs.New(domain, ErrEmpty, ErrURL)
	ErrExhaustedBackoff = errs.New(domain, ErrExhausted, ErrBackoff)
)

const (
	defaultTimeout = 30 * time.Second
	minBackoff     = 125 * time.Millisecond
	maxBackoff     = 10 * time.Second
)

// New issues an HTTP GET request based on the input URL `url` and timeout `timeout`
//
// The body of the response is then used as an audio stream
func New(logger logx.Logger, url string, timeout time.Duration) (*http.Response, context.CancelFunc, error) {
	if len(url) > 0 {
		if url[0] == '"' {
			url = url[1:]
		}
		if url[len(url)-1] == '"' {
			url = url[:len(url)-1]
		}
	}

	if url == "" {
		return nil, nil, ErrEmptyURL
	}

	res, done, err := doWithBackoff(logger, url, timeout)
	if err != nil {
		return nil, nil, err
	}

	return res, done, nil
}

func doWithBackoff(logger logx.Logger, url string, timeout time.Duration) (*http.Response, context.CancelFunc, error) {
	backoff := minBackoff

	for {
		res, done, err := do(url, timeout)
		if err == nil {
			return res, done, nil
		}

		logger.Warn("request has failed",
			attr.String("error", err.Error()),
			attr.String("backoff", backoff.String()),
		)

		time.Sleep(backoff)

		backoff = backoff * 2

		if backoff > maxBackoff {
			return res, done, fmt.Errorf("%w: %w", ErrExhaustedBackoff, err)
		}
	}
}

func do(url string, timeout time.Duration) (*http.Response, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		cancel()

		return nil, nil, err
	}

	req.Header.Set("Content-Type", "audio/wav")

	res, err := (&http.Client{
		Timeout: defaultTimeout,
	}).Do(req)

	if err != nil {
		cancel()

		return nil, nil, err
	}

	return res, cancel, nil
}
