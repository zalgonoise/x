package ghttp

import (
	"context"
	"net/http"
)

// Handler is a basic path-to-HandlerFunc pair
type Handler struct {
	Path       string
	Fn         http.HandlerFunc
	Middleware []MiddlewareFn
}

// NewHandler joins a path with a HandlerFunc, returning a Handler
func NewHandler(path string, fn http.HandlerFunc, middleware ...MiddlewareFn) Handler {
	return Handler{
		Path:       path,
		Fn:         fn,
		Middleware: middleware,
	}
}

// ParseFn is a function that converts a HTTP request into a request object of the caller's choice
type ParseFn[Q any] func(ctx context.Context, r *http.Request) (*Q, error)

// RouteFn is a function that converts a HTTP request into a HandlerFunc (appropriate to this type of request)
type RouteFn func(w http.ResponseWriter, r *http.Request) http.HandlerFunc

// QueryFn is a function that executes an action based on the input context `ctx` and query object `query`,
// and returns the HTTP response's status, message, body object `answer`, a headers map and an error
type QueryFn[Q any, A any] func(ctx context.Context, query *Q) (status int, msg string, answer *A, headers map[string]string, err error)

// ExecFn is a function that executes an action based on the input context `ctx` and query object `query`,
// and returns the HTTP response's status, message, a headers map and an error
type ExecFn[Q any] func(ctx context.Context, query *Q) (status int, msg string, headers map[string]string, err error)

// MiddlewareFn is a function that wraps a http.HandlerFunc, as HTTP middleware
type MiddlewareFn func(next http.HandlerFunc) http.HandlerFunc

// Query is a generic function that creates a HandlerFunc which will take in a context and a query object, and returns
// a HTTP status, a response message, a response object and an error
func Query[Q any, A any](name string, parseFn ParseFn[Q], queryFn QueryFn[Q, A]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, s := NewCtxAndSpan(r, name)
		defer s.End()

		if queryFn == nil {
			panic("a query function must be specified")
		}

		var query = new(Q)
		var err error

		if parseFn != nil {
			query, err = parseFn(ctx, r)
			if err != nil {
				ErrResponse(400, "failed to parse request", err, nil).WriteHTTP(ctx, w)
				return
			}
		}

		status, msg, answer, headers, err := queryFn(ctx, query)
		if err != nil {
			if status < 400 {
				status = 500
			}
			if msg == "" {
				msg = "operation failed"
			}
			ErrResponse(status, msg, err, headers).WriteHTTP(ctx, w)
			return
		}

		if status > 399 {
			status = 200
		}
		OKResponse(status, msg, answer, headers).WriteHTTP(ctx, w)
	}
}

// Exec is a generic function that creates a HandlerFunc which will take in a context and a query object, and returns
// a HTTP status, a response message and an error
func Exec[Q any](name string, parseFn ParseFn[Q], execFn ExecFn[Q]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, s := NewCtxAndSpan(r, name)
		defer s.End()

		if execFn == nil {
			panic("an exec function must be specified")
		}

		var query = new(Q)
		var err error

		if parseFn != nil {
			query, err = parseFn(ctx, r)
			if err != nil {
				ErrResponse(400, "failed to parse request", err, nil).WriteHTTP(ctx, w)
				return
			}
		}

		status, msg, headers, err := execFn(ctx, query)
		if err != nil {
			if status < 400 {
				status = 500
			}
			if msg == "" {
				msg = "operation failed"
			}
			ErrResponse(status, msg, err, headers).WriteHTTP(ctx, w)
			return
		}
		if status > 399 {
			status = 200
		}

		OKResponse[Q](status, msg, nil, headers).WriteHTTP(ctx, w)
	}
}
