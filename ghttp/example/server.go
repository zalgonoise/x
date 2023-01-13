package main

import (
	"context"
	"net/http"

	"github.com/zalgonoise/x/ghttp"
)

type User struct {
	ID       int
	Name     string
	Username string
}

type server struct {
	http  ghttp.Server
	users map[int]*User
}

func newServer() *server {
	return &server{
		users: map[int]*User{
			0: {ID: 0, Name: "Me", Username: "me"},
			1: {ID: 1, Name: "The other", Username: "the_other"},
			2: {ID: 2, Name: "Someone else", Username: "someone_else"},
		},
	}
}

func (s *server) usersHandler() []ghttp.Handler {
	p := "/users/"

	return []ghttp.Handler{
		{
			Method: http.MethodGet,
			Path:   p,
			Fn:     s.usersGetListRoute(p),
		},
		{
			Method: http.MethodPost,
			Path:   p,
			Fn:     s.usersCreate(),
		},
		{
			Method: http.MethodDelete,
			Path:   p,
			Fn:     s.usersDelete(),
		},
		{
			Method: http.MethodPut,
			Path:   p,
			Fn:     s.usersUpdate(),
		},
	}
}

func (s *server) endpoints() ghttp.Endpoints {
	e := ghttp.NewEndpoints()
	e.Set(s.usersHandler()...)
	return e
}

func main() {
	srv := newServer()
	srv.http = ghttp.NewServer(srv.endpoints(), 8080)

	err := srv.http.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
