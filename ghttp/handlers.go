package ghttp

import (
	"context"
	"net/http"

	"github.com/zalgonoise/attr"
)

type router struct {
	Path   string
	Routes map[string]http.HandlerFunc
}

// ServeHTTP implements the http.Handler interface
func (m *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := m.Routes[r.Method]; ok {
		handler(w, r)
		return
	}
	http.NotFound(w, r)
}

// NewMux creates a http mux based off of a set of Handlers that share the
// same path, but present different HandlerFuncs for different HTTP methods
//
// Middleware funcs are applied to the HandlerFunc, in the Routes map, in the mux.
func NewRouter(handlers ...Handler) http.Handler {
	if len(handlers) == 0 {
		return nil
	}
	p := handlers[0].Path
	m := &router{
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
// and returns a pointer to a Response for the answer type
type ExecFn[Q any, A any] func(ctx context.Context, query *Q) *Response[A]

// MiddlewareFn is a function that wraps a http.HandlerFunc, as HTTP middleware
type MiddlewareFn func(next http.HandlerFunc) http.HandlerFunc

// Do is a generic function that creates a HandlerFunc which will take in a context and a query object, and returns
// a HTTP status, a response message, a response object and an error
func Do[Q any, A any](name string, parseFn ParseFn[Q], queryFn ExecFn[Q, A]) http.HandlerFunc {
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
				s.Event("failed to parse the request", attr.String("error", err.Error()))
				NewResponse[A](http.StatusBadRequest, err.Error()).WriteHTTP(ctx, w)
				return
			}
		}

		res := queryFn(ctx, query)
		if res.Status > 399 {
			s.Event("operation error", attr.Int("status", res.Status), attr.String("error", res.Message))
			res.WriteHTTP(ctx, w)
			return
		}

		s.Event("operation successful", attr.Int("status", res.Status))
		res.WriteHTTP(ctx, w)
	}
}
