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

func (r ReaderHook) Read(p []byte) (n int, err error) {
	if err = r.fn(r.h, p); err != nil {
		return 0, err
	}

	return r.Reader.Read(p)
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

func (r ReaderContextHook) Read(p []byte) (n int, err error) {
	if err = r.fn(r.ctx, r.h, p); err != nil {
		return 0, err
	}

	return r.Reader.Read(p)
}

func NewReaderContextHook(ctx context.Context, h *Header, r io.Reader, fn HookContextFunc) ReaderContextHook {
	return ReaderContextHook{
		h:      h,
		Reader: r,
		ctx:    ctx,
		fn:     fn,
	}
}