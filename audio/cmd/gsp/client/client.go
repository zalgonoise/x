package client

import (
	"context"
	"net/http"
	"time"
)

type err string

func (e err) Error() string { return (string)(e) }

const (
	ErrEmptyURL err = "gsp/client: empty URL"
)

type streamClient struct {
	client *http.Client
	req    *http.Request
	ctx    context.Context
}

type Doer interface {
	Do() (*http.Response, error)
	Context() context.Context
}

func (c *streamClient) Do() (*http.Response, error) {
	if c.client == nil {
		return http.DefaultClient.Do(c.req)
	}
	return c.client.Do(c.req)
}

func (c *streamClient) Context() context.Context {
	return c.ctx
}

func New(url string, timeout *time.Duration) (Doer, context.CancelFunc, error) {
	if url == "" {
		return nil, nil, ErrEmptyURL
	}
	ctx, cancel := context.WithCancel(context.Background())
	if timeout != nil {
		ctx, cancel = context.WithTimeout(ctx, *timeout)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "audio/wav")

	doer := new(streamClient)
	doer.req = req
	doer.ctx = ctx
	return doer, cancel, nil
}

func WithClient(url string, timeout *time.Duration, client *http.Client) (Doer, context.CancelFunc, error) {
	doer, cancel, err := New(url, timeout)
	if err != nil {
		return nil, nil, err
	}

	doer.(*streamClient).client = client
	return doer, cancel, err
}
