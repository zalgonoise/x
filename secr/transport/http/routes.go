package http

import (
	"net/http"
	"strings"

	"github.com/zalgonoise/x/ghttp"
)

func getPath(path string) []string {
	splitPath := strings.Split(path, "/")
	var out = make([]string, 0, len(splitPath))
	for _, item := range splitPath {
		if item != "" && item != " " && item != "\n" && item != "\t" {
			out = append(out, item)
		}
	}
	return out
}

const userPath = "users"
const secrPath = "secrets"

func (s *server) endpoints() ghttp.Endpoints {
	e := ghttp.NewEndpoints()
	e.Set(s.usersHandler()...)
	e.Set(s.sessionsHandler()...)
	return e
}

func (s *server) sessionsHandler() []ghttp.Handler {
	return []ghttp.Handler{
		{
			Method: http.MethodPost,
			Path:   "/login",
			Fn:     s.login(),
		},
		{
			Method: http.MethodPost,
			Path:   "/logout",
			Fn:     s.logout(),
		},
		{
			Method: http.MethodPost,
			Path:   "/recover",
			Fn:     s.changePassword(),
		},
		{
			Method: http.MethodPost,
			Path:   "/refresh",
			Fn:     s.refresh(),
		},
	}
}

func (s *server) usersHandler() []ghttp.Handler {
	p := "/users/"

	return []ghttp.Handler{
		{
			Method: http.MethodGet,
			Path:   p,
			Fn:     s.usersGetRoute(),
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
		{
			Method: http.MethodPost,
			Path:   p,
			Fn:     s.usersPostRoute(),
		},
		{
			Method: http.MethodDelete,
			Path:   p,
			Fn:     s.usersDeleteRoute(),
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
		{
			Method: http.MethodPut,
			Path:   p,
			Fn:     s.usersUpdate(),
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
	}
}

func (s *server) usersGetRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		splitPath := getPath(r.URL.Path)
		switch len(splitPath) {
		case 1:
			if splitPath[0] == userPath {
				s.usersList()(w, r)
				return
			}
		case 2:
			if splitPath[0] == userPath {
				s.usersGet()(w, r)
				return
			}
		case 3:
			if splitPath[0] == userPath && splitPath[2] == secrPath {
				s.secretsList()(w, r)
				return
			}
		case 4:
			if splitPath[0] == userPath && splitPath[2] == secrPath {
				s.secretsGet()(w, r)
				return
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}

func (s *server) usersPostRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		splitPath := getPath(r.URL.Path)

		switch len(splitPath) {
		case 1:
			if splitPath[0] == userPath {
				s.usersCreate()(w, r)
				return
			}
		case 3:
			if splitPath[0] == userPath && splitPath[2] == secrPath {
				s.WithAuth()(s.secretsCreate())(w, r)
				return
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}

func (s *server) usersDeleteRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		splitPath := getPath(r.URL.Path)

		switch len(splitPath) {
		case 1:
			if splitPath[0] == userPath {
				s.usersDelete()(w, r)
				return
			}
		case 3:
			if splitPath[0] == userPath && splitPath[2] == secrPath {
				s.secretsDelete()(w, r)
				return
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}
