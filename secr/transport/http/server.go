package http

import (
	"context"
	"fmt"
	stdhttp "net/http"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
)

type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type server struct {
	ep   API
	port int
	srv  *stdhttp.Server
}

func NewServer(api API, port int) Server {
	mux := stdhttp.NewServeMux()
	httpSrv := &stdhttp.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: mux,
	}
	srv := server{
		ep:   api,
		port: port,
		srv:  httpSrv,
	}

	// handlers
	// TODO: implement dynamic URL paths (e.g.: /users/me/secrets/github)
	mux.HandleFunc("/login", srv.ep.Login)
	mux.HandleFunc("/logout", srv.ep.Logout)
	mux.HandleFunc("/recover", srv.ep.ChangePassword)
	mux.HandleFunc("/refresh", srv.ep.Refresh)
	mux.HandleFunc("/users/create", srv.ep.CreateUser)
	mux.HandleFunc("/users/get", srv.ep.GetUser)
	mux.HandleFunc("/users", srv.ep.ListUsers)
	mux.HandleFunc("/users/update", srv.ep.UpdateUser)
	mux.HandleFunc("/users/delete", srv.ep.DeleteUser)
	mux.HandleFunc("/secrets/create", srv.ep.CreateSecret)
	mux.HandleFunc("/secrets/get", srv.ep.GetSecret)
	mux.HandleFunc("/secrets", srv.ep.ListSecrets)
	mux.HandleFunc("/secrets/delete", srv.ep.DeleteSecret)
	return srv
}

func (s server) Start(ctx context.Context) error {
	_, span := spanner.Start(ctx, "http.Start")
	defer span.End()

	err := s.srv.ListenAndServe()
	if err != nil {
		span.Event("failed to start HTTP server", attr.New("error", err.Error()))
		return err
	}
	return nil
}

func (s server) Stop(ctx context.Context) error {
	ctx, span := spanner.Start(ctx, "http.Stop")
	defer span.End()

	err := s.srv.Shutdown(ctx)
	if err != nil {
		span.Event("failed to stop HTTP server", attr.New("error", err.Error()))
		return err
	}
	return nil
}
