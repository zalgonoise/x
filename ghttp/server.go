package ghttp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
)

// Server describes the functionalities of a HTTP server, once spawned
type Server interface {
	// Start initializes the HTTP server, returning an error. This is a blocking call
	Start(ctx context.Context) error
	// Stop gracefully shuts-down the HTTP server, returning an error
	Stop(ctx context.Context) error
}

type server struct {
	endpoints Endpoints
	port      int
	srv       *http.Server
}

// NewServer spawns a new HTTP server for port `port`, configured with Endpoints
// `endpoints`
func NewServer(endpoints Endpoints, port int) Server {
	mux := http.NewServeMux()
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: mux,
	}
	srv := &server{
		endpoints: endpoints,
		port:      port,
		srv:       httpSrv,
	}

	for path, handlers := range endpoints.E {
		m := NewMux(handlers...)
		mux.HandleFunc(path, m.ServeHTTP)
	}

	return srv
}

// Start initializes the HTTP server, returning an error. This is a blocking call
func (s *server) Start(ctx context.Context) error {
	_, span := spanner.Start(ctx, "http.Start")
	defer span.End()

	err := s.srv.ListenAndServe()
	if err != nil {
		span.Event("failed to start HTTP server", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// Stop gracefully shuts-down the HTTP server, returning an error
func (s *server) Stop(ctx context.Context) error {
	ctx, span := spanner.Start(ctx, "http.Stop")
	defer span.End()

	err := s.srv.Shutdown(ctx)
	if err != nil {
		span.Event("failed to stop HTTP server", attr.New("error", err.Error()))
		return err
	}
	return nil
}
