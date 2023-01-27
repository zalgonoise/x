package http

import (
	"net/http"

	"github.com/zalgonoise/x/ghttp"
)

const userPath = "users"
const secrPath = "secrets"
const shareAction = "share"
const sharePath = "shares"

func (s *server) endpoints() ghttp.Endpoints {
	e := ghttp.NewEndpoints()
	e.Set(s.usersHandler()...)
	e.Set(s.secretsHandler()...)
	e.Set(s.sharesHandler()...)
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
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
		{
			Method: http.MethodPost,
			Path:   "/recover",
			Fn:     s.changePassword(),
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
		{
			Method: http.MethodPost,
			Path:   "/refresh",
			Fn:     s.refresh(),
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
	}
}

func (s *server) secretsHandler() []ghttp.Handler {
	p := "/secrets/"
	return []ghttp.Handler{
		{
			Method: http.MethodGet,
			Path:   p,
			Fn:     s.secretsGetRoute(),
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
		{
			Method: http.MethodPost,
			Path:   p,
			Fn:     s.secretsCreateRoute(),
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
		{
			Method: http.MethodDelete,
			Path:   p,
			Fn:     s.secretsDelete(),
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
	}
}

func (s *server) secretsGetRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		splitPath := getPath(r.URL.Path)
		switch len(splitPath) {
		case 1:
			if splitPath[0] == secrPath {
				s.secretsList()(w, r)
				return
			}
		case 2:
			if splitPath[0] == secrPath {
				s.secretsGet()(w, r)
				return
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}

func (s *server) secretsCreateRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		splitPath := getPath(r.URL.Path)
		switch len(splitPath) {
		case 1:
			if splitPath[0] == secrPath {
				s.secretsCreate()(w, r)
				return
			}
		case 3:
			if splitPath[0] == secrPath && splitPath[2] == shareAction {
				s.sharesCreate()(w, r)
				return
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}

func (s *server) usersHandler() []ghttp.Handler {
	p := "/users/"
	return []ghttp.Handler{
		{
			Method: http.MethodPost,
			Path:   p,
			Fn:     s.usersCreate(),
		},
		{
			Method: http.MethodGet,
			Path:   p,
			Fn:     s.usersGetRoute(),
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
		{
			Method: http.MethodDelete,
			Path:   p,
			Fn:     s.usersDelete(),
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
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}

func (s *server) sharesHandler() []ghttp.Handler {
	p := "/shares/"
	return []ghttp.Handler{
		{
			Method: http.MethodGet,
			Path:   p,
			Fn:     s.sharesGetRoute(),
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
		{
			Method: http.MethodDelete,
			Path:   p,
			Fn:     s.sharesDelete(),
			Middleware: []ghttp.MiddlewareFn{
				s.WithAuth(),
			},
		},
	}
}

func (s *server) sharesGetRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		splitPath := getPath(r.URL.Path)
		switch len(splitPath) {
		case 1:
			if splitPath[0] == sharePath {
				s.sharesList()(w, r)
				return
			}
		case 2:
			if splitPath[0] == sharePath {
				s.sharesGet()(w, r)
				return
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}
