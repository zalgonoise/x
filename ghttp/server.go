package ghttp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
)

const defaultPort = 8080

type Server struct {
	*http.Server
}

// NewServer spawns a default HTTP server on port `port`, configured with endpoints `endpoints`
func NewServer(port int, endpoints Endpoints) *Server {
	if port <= 0 {
		port = defaultPort
	}
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr: fmt.Sprintf(":%v", port),
	}
	return WithServerAndMux(srv, mux, endpoints)
}

// WithServer wraps a standard library *http.Server with the input endpoints `endpoints`. The caller
// must configure the server's address; otherwise it defaults to ":8080"
func WithServer(srv *http.Server, endpoints Endpoints) *Server {
	if srv == nil {
		return NewServer(defaultPort, endpoints)
	}
	if srv.Addr == "" {
		srv.Addr = fmt.Sprintf(":%v", defaultPort)
	}
	if srv.Handler == nil {
		return WithServerAndMux(srv, http.NewServeMux(), endpoints)
	}

	if mux, ok := srv.Handler.(interface {
		HandleFunc(string, func(http.ResponseWriter, *http.Request))
	}); ok {
		WithServerAndMux(srv, mux, endpoints)
	}

	return WithServerAndMux(srv, http.NewServeMux(), endpoints)
}

// WithMux creates a HTTP server based on the input multiplexer. The input `mux` has to implement
// the HandleFunc(string, func(http.ResponseWriter, *http.Request)) method
func WithMux(
	port int,
	mux interface {
		HandleFunc(string, func(http.ResponseWriter, *http.Request))
	},
	endpoints Endpoints,
) *Server {
	if port <= 0 {
		port = defaultPort
	}
	if mux == nil {
		return NewServer(port, endpoints)
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: mux.(http.Handler),
	}
	return WithServerAndMux(srv, mux, endpoints)
}

// WithServerAndMux wraps existing *http.Server and a mux, with the input endpoints `endpoints`.
// The input `mux` has to implement the HandleFunc(string, func(http.ResponseWriter, *http.Request))
// method
//
// The input server `srv`'s Handler will be replaced with the input mux `mux`
func WithServerAndMux(
	srv *http.Server,
	mux interface {
		HandleFunc(string, func(http.ResponseWriter, *http.Request))
	},
	endpoints Endpoints,
) *Server {
	if srv == nil {
		return WithMux(defaultPort, mux, endpoints)
	}
	if srv.Addr == "" {
		srv.Addr = fmt.Sprintf(":%v", defaultPort)
	}
	for path, handlers := range endpoints {
		r := NewRouter(handlers...)
		mux.HandleFunc(path, r.ServeHTTP)
	}
	srv.Handler = mux.(http.Handler)

	return &Server{srv}
}

// Start initializes the HTTP server, returning an error. This is a blocking call
func (s *Server) Start(ctx context.Context) error {
	_, span := spanner.Start(ctx, "http.Start")
	defer span.End()

	err := s.ListenAndServe()
	if err != nil {
		span.Event("failed to start HTTP server", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// Stop gracefully shuts-down the HTTP server, returning an error
func (s *Server) Stop(ctx context.Context) error {
	ctx, span := spanner.Start(ctx, "http.Stop")
	defer span.End()

	err := s.Shutdown(ctx)
	if err != nil {
		span.Event("failed to stop HTTP server", attr.New("error", err.Error()))
		return err
	}
	return nil
}
