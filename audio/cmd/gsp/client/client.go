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

// Doer interface describes an object with a http.Client with a http.Request
// and a context.Context.
//
// The exposed methods allow issuing the http.Request, as well as one to
// retrieve its set context.Context
type Doer interface {
	// Do issues the http.Request in the Doer, returning a http.Response and an error
	Do() (*http.Response, error)
	// Context returns the set context.Context
	Context() context.Context
}

// Do issues the http.Request in the Doer, returning a http.Response and an error
func (c *streamClient) Do() (*http.Response, error) {
	if c.client == nil {
		return http.DefaultClient.Do(c.req)
	}
	return c.client.Do(c.req)
}

// Context returns the set context.Context
func (c *streamClient) Context() context.Context {
	return c.ctx
}

// New creates a basic Doer interface, based on the input URL `url` and timeout `timeout`
//
// This Doer is based on the default http.Client from Go's standard library
func New(url string, timeout *time.Duration) (Doer, context.CancelFunc, error) {
	if url == "" {
		return nil, nil, ErrEmptyURL
	}
	if len(url) > 0 {
		if url[0] == '"' {
			url = url[1:]
		}
		if url[len(url)-1] == '"' {
			url = url[:len(url)-1]
		}
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

// WithClient is just like New, but it is configured with the input
// http.Client `client`, instead
func WithClient(url string, timeout *time.Duration, client *http.Client) (Doer, context.CancelFunc, error) {
	doer, cancel, err := New(url, timeout)
	if err != nil {
		return nil, nil, err
	}

	doer.(*streamClient).client = client
	return doer, cancel, err
}
