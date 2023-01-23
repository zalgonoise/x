package http

import (
	"context"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/service"
)

// Server describes the general actions the HTTP server will expose
type Server interface {
	// Start will initialize the HTTP server, returning an error. This is a blocking call.
	Start(ctx context.Context) error
	// Stop will gracefully shut-down the HTTP server, returning an error
	Stop(ctx context.Context) error
}

type server struct {
	s    service.Service
	HTTP *ghttp.Server
}

// NewServer creates a new HTTP server from the input port `port` and service `s`
func NewServer(port int, s service.Service) Server {
	srv := &server{s: s}
	srv.HTTP = ghttp.NewServer(port, srv.endpoints())
	return srv
}

// Start will initialize the HTTP server, returning an error. This is a blocking call.
func (s *server) Start(ctx context.Context) error {
	_, span := spanner.Start(ctx, "http.Start")
	defer span.End()

	err := s.HTTP.ListenAndServe()
	if err != nil {
		span.Event("failed to start HTTP server", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// Stop will gracefully shut-down the HTTP server, returning an error
func (s *server) Stop(ctx context.Context) error {
	ctx, span := spanner.Start(ctx, "http.Stop")
	defer span.End()

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		span.Event("failed to stop HTTP server", attr.New("error", err.Error()))
		return err
	}
	return nil
}
