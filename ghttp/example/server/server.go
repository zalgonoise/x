package server

import (
	"github.com/zalgonoise/x/ghttp"
)

type User struct {
	ID       int
	Name     string
	Username string
}

type Server struct {
	HTTP  *ghttp.Server
	users map[int]*User
}

func New() *Server {
	return &Server{
		users: map[int]*User{
			0: {ID: 0, Name: "Me", Username: "me"},
			1: {ID: 1, Name: "The other", Username: "the_other"},
			2: {ID: 2, Name: "Someone else", Username: "someone_else"},
		},
	}
}

func (s *Server) Endpoints() ghttp.Endpoints {
	e := ghttp.NewEndpoints()
	e.Set(s.usersHandler()...)
	return e
}
