package ghttp

import (
	"context"
	"net/http"
)

type mux struct {
	Path   string
	Routes map[string]http.HandlerFunc
}

func (m *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := m.Routes[r.Method]; ok {
		handler(w, r)
		return
	}
	http.NotFound(w, r)
}

func NewMux(handlers ...Handler) http.Handler {
	if len(handlers) == 0 {
		return nil
	}
	p := handlers[0].Path
	m := &mux{
		Path:   p,
		Routes: make(map[string]http.HandlerFunc),
	}

	for _, h := range handlers {
		if h.Path != p {
			continue
		}
		var fn = h.Fn
		for i := len(h.Middleware) - 1; i >= 0; i-- {
			fn = h.Middleware[i](fn)
		}

		m.Routes[h.Method] = fn
	}
	return m
}

// Handler is a basic path-to-HandlerFunc pair
type Handler struct {
	Method     string
	Path       string
	Fn         http.HandlerFunc
	Middleware []MiddlewareFn
}

// NewHandler joins a path with a HandlerFunc, returning a Handler
func NewHandler(method, path string, fn http.HandlerFunc, middleware ...MiddlewareFn) Handler {
	return Handler{
		Method:     method,
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
type QueryFn[Q any, A any] func(ctx context.Context, query *Q) *Response[A]

// ExecFn is a function that executes an action based on the input context `ctx` and query object `query`,
// and returns the HTTP response's status, message, a headers map and an error
type ExecFn[Q any] func(ctx context.Context, query *Q) *Response[Q]

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
				res := NewResponse[A](http.StatusBadRequest, "failed to parse request")
				res.Error = err.Error()
				res.WriteHTTP(ctx, w)
				return
			}
		}

		queryFn(ctx, query).WriteHTTP(ctx, w)
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
				res := NewResponse[Q](http.StatusBadRequest, "failed to parse request")
				res.Error = err.Error()
				res.WriteHTTP(ctx, w)
				return
			}
		}

		execFn(ctx, query).WriteHTTP(ctx, w)
	}
}
