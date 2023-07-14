package http

import (
	"context"
	"net/http"
	"time"

	"github.com/zalgonoise/x/audio/errs"
)

const (
	domain   = errs.Domain("audio/http")
	ErrEmpty = errs.Kind("empty")
	ErrURL   = errs.Entity("URL")
)

var ErrEmptyURL = errs.New(domain, ErrEmpty, ErrURL)

const defaultTimeout = 30 * time.Second

// New issues an HTTP GET request based on the input URL `url` and timeout `timeout`
//
// The body of the response is then used as an audio stream
func New(url string, timeout time.Duration) (*http.Response, context.CancelFunc, error) {
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
