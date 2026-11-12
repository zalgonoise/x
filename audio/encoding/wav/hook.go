package wav

import (
	"context"
	"io"
)

type HookFunc func(*Header, []byte) error

type ReaderHook struct {
	h      *Header
	Reader io.Reader

	fn HookFunc
}

// Read implements the io.Reader interface. The hook sees exactly the bytes read (p[:n]), including bytes returned
// together with an error such as io.EOF; a hook error takes precedence over the read error.
func (r ReaderHook) Read(p []byte) (n int, err error) {
	if r.Reader == nil {
		return 0, io.EOF
	}

	n, err = r.Reader.Read(p)
	if n > 0 && r.fn != nil {
		if hookErr := r.fn(r.h, p[:n]); hookErr != nil {
			return n, hookErr
		}
	}

	return n, err
}

func NewReaderHook(h *Header, r io.Reader, fn HookFunc) ReaderHook {
	return ReaderHook{
		h:      h,
		Reader: r,
		fn:     fn,
	}
}

type HookContextFunc func(context.Context, *Header, []byte) error

type ReaderContextHook struct {
	h      *Header
	Reader io.Reader

	ctx context.Context
	fn  HookContextFunc
}

// Read implements the io.Reader interface. The hook sees exactly the bytes read (p[:n]), including bytes returned
// together with an error such as io.EOF; a hook error takes precedence over the read error.
func (r ReaderContextHook) Read(p []byte) (n int, err error) {
	if r.Reader == nil {
		return 0, io.EOF
	}

	n, err = r.Reader.Read(p)
	if n > 0 && r.fn != nil {
		if hookErr := r.fn(r.ctx, r.h, p[:n]); hookErr != nil {
			return n, hookErr
		}
	}

	return n, err
}

func NewReaderContextHook(ctx context.Context, h *Header, r io.Reader, fn HookContextFunc) ReaderContextHook {
	return ReaderContextHook{
		h:      h,
		Reader: r,
		ctx:    ctx,
		fn:     fn,
	}
}
