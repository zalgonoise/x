package http

import (
	"context"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/service"
)

type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type server struct {
	s    service.Service
	HTTP *ghttp.Server
}

func NewServer(s service.Service, port int) Server {
	srv := server{}
	srv.HTTP = ghttp.NewServer(port, srv.endpoints())
	return srv
}

func (s server) Start(ctx context.Context) error {
	_, span := spanner.Start(ctx, "http.Start")
	defer span.End()

	err := s.HTTP.ListenAndServe()
	if err != nil {
		span.Event("failed to start HTTP server", attr.New("error", err.Error()))
		return err
	}
	return nil
}

func (s server) Stop(ctx context.Context) error {
	ctx, span := spanner.Start(ctx, "http.Stop")
	defer span.End()

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		span.Event("failed to stop HTTP server", attr.New("error", err.Error()))
		return err
	}
	return nil
}
