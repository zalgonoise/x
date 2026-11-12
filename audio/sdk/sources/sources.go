// Package sources provides sdk.Source implementations: local files, a single io.Reader (such as stdin) and HTTP.
package sources

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

// Errors returned by the sources.
var (
	// ErrConsumed is returned by a Reader source opened a second time: its stream can only be read once.
	ErrConsumed = errors.New("sources: reader already consumed")
	// ErrUnexpectedStatus is returned by an HTTP source for a response outside 2xx.
	ErrUnexpectedStatus = errors.New("sources: unexpected HTTP status")
	// ErrInvalidURL is returned by NewHTTP for a URL that is not absolute http(s).
	ErrInvalidURL = errors.New("sources: invalid URL")
)

// File is a Source reading a local file, opened anew on every Open.
type File struct{ path string }

// NewFile returns a Source for the file at path. The file is opened by Open, so a missing file is reported there.
func NewFile(path string) *File { return &File{path: path} }

// Open implements sdk.Source.
func (f *File) Open(context.Context) (io.ReadCloser, error) { return os.Open(f.path) }

// Reader is a Source over an io.Reader that can be read once, such as stdin.
//
// Closing the stream it opens always unblocks a pending Read, even on a descriptor whose read(2) cannot be interrupted
// (stdin from a terminal or a shell pipe): the Reader copies from r on its own goroutine through a pipe, and Close
// closes the pipe. That goroutine then exits when its read returns; if r is an io.Closer it is closed too, which ends
// the read at once where the platform allows it.
type Reader struct {
	mu   sync.Mutex
	r    io.Reader
	used bool
}

// NewReader returns a single-use Source for r.
func NewReader(r io.Reader) *Reader { return &Reader{r: r} }

// Open implements sdk.Source. A second call returns ErrConsumed.
func (s *Reader) Open(context.Context) (io.ReadCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.used {
		return nil, ErrConsumed
	}

	s.used = true

	pr, pw := io.Pipe()

	go func() {
		_, err := io.Copy(pw, s.r)
		pw.CloseWithError(err) // nil means io.EOF for the reading side
	}()

	return &pipeStream{PipeReader: pr, src: s.r}, nil
}

// pipeStream is the stream of a Reader source: closing it unblocks a pending Read at once.
type pipeStream struct {
	*io.PipeReader

	src io.Reader
}

func (p *pipeStream) Close() error {
	err := p.PipeReader.Close()

	if c, ok := p.src.(io.Closer); ok {
		err = errors.Join(err, c.Close())
	}

	return err
}

// HTTP is a Source that GETs a URL on every Open: one request per Run, so a Pipeline can reconnect by running again.
type HTTP struct {
	url    string
	client *http.Client
	header http.Header
}

// HTTPOption configures an HTTP source.
type HTTPOption func(*HTTP)

// WithClient sets the HTTP client. The default has no overall timeout, which would cut endless streams, and bounds
// each phase instead: 10 s to connect, 10 s for the TLS handshake and 15 s for the response headers.
func WithClient(c *http.Client) HTTPOption {
	return func(h *HTTP) {
		if c != nil {
			h.client = c
		}
	}
}

// WithHeader adds a request header.
func WithHeader(key, value string) HTTPOption {
	return func(h *HTTP) { h.header.Add(key, value) }
}

// NewHTTP returns a Source for an absolute http or https URL.
func NewHTTP(rawURL string, opts ...HTTPOption) (*HTTP, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidURL, rawURL)
	}

	h := &HTTP{url: u.String(), client: defaultClient(), header: http.Header{"Accept": {"audio/*, */*;q=0.5"}}}
	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

// Timeouts of the default HTTP client: each phase is bounded, the body is not (live streams never end).
const (
	dialTimeout           = 10 * time.Second
	keepAlive             = 30 * time.Second
	tlsHandshakeTimeout   = 10 * time.Second
	responseHeaderTimeout = 15 * time.Second
	idleConnTimeout       = 90 * time.Second
)

func defaultClient() *http.Client {
	dialer := &net.Dialer{Timeout: dialTimeout, KeepAlive: keepAlive}

	// streams are long-lived and each Open is a new request: reusing connections gains nothing, and keeping them would
	// leave idle-connection goroutines behind every source
	return &http.Client{Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		DisableKeepAlives:     true,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ResponseHeaderTimeout: responseHeaderTimeout,
		IdleConnTimeout:       idleConnTimeout,
	}}
}

// Open implements sdk.Source. A response outside 2xx is closed and reported as ErrUnexpectedStatus. Closing the
// returned body, or cancelling ctx, aborts the request.
func (h *HTTP) Open(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header = h.header.Clone()

	res, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		_ = res.Body.Close()

		return nil, fmt.Errorf("%w: %s", ErrUnexpectedStatus, res.Status)
	}

	return res.Body, nil
}
