package ghttp

import (
	"context"
	"net/http"
)

// Handler is a basic path-to-HandlerFunc pair
type Handler struct {
	Path string
	Fn   http.HandlerFunc
}

// NewHandler joins a path with a HandlerFunc, returning a Handler
func NewHandler(path string, fn http.HandlerFunc) Handler {
	return Handler{
		Path: path,
		Fn:   fn,
	}
}

// Query is a generic function that creates a HandlerFunc which will take in a context and a query object, and returns
// a HTTP status, a response message, a response object and an error
func Query[Q any, A any](name string, queryFn func(ctx context.Context, query *Q) (int, string, *A, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, s := NewCtxAndSpan(r, name)
		defer s.End()

		query, err := ReadBody[Q](ctx, r)
		if err != nil {
			res := ErrResponse(400, "failed to read request from body", err)
			res.WriteHTTP(ctx, w)
			return
		}

		status, msg, answer, err := queryFn(ctx, query)
		if err != nil {
			if status < 400 {
				status = 500
			}
			res := ErrResponse(status, "operation failed", err)
			res.WriteHTTP(ctx, w)
			return
		}
		if status > 399 {
			status = 200
		}

		res := OKResponse(200, msg, answer)
		res.WriteHTTP(ctx, w)
	}
}

// Exec is a generic function that creates a HandlerFunc which will take in a context and a query object, and returns
// a HTTP status, a response message and an error
func Exec[Q any](name string, queryFn func(ctx context.Context, query *Q) (int, string, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, s := NewCtxAndSpan(r, "http.AddRecord")
		defer s.End()

		query, err := ReadBody[Q](ctx, r)
		if err != nil {
			res := ErrResponse(400, "failed to read request from body", err)
			res.WriteHTTP(ctx, w)
			return
		}

		status, msg, err := queryFn(ctx, query)
		if err != nil {
			if status < 400 {
				status = 500
			}
			res := ErrResponse(status, "operation failed", err)
			res.WriteHTTP(ctx, w)
			return
		}
		if status > 399 {
			status = 200
		}

		res := OKResponse[Q](200, msg, nil)
		res.WriteHTTP(ctx, w)
	}
}
