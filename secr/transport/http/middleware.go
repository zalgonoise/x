package http

import (
	"net/http"

	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/authz"
)

func (s *server) WithAuth() ghttp.MiddlewareFn {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			var username string
			splitPath := getPath(r.URL.Path)

			if len(splitPath) > 1 {
				username = splitPath[1]
			} else {
				next(w, r)
				return
			}
			if username == "" {
				http.Error(w, "no username provided", http.StatusBadRequest)
				return
			}

			if token, ok := getToken(r); ok {
				ok, err := s.s.Validate(r.Context(), username, token)
				if err == nil && ok {
					// wrap caller info in context
					next(w, authz.SignRequest(username, r))
					return
				}
			}

			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}
