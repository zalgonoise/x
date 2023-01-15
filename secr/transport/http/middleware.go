package http

import (
	"net/http"
	"strings"

	"github.com/zalgonoise/x/ghttp"
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

			token := r.Header.Get("Authorization")
			if token != "" {
				t := strings.TrimPrefix(token, "Bearer ")
				if t != "" {
					ok, err := s.s.Validate(r.Context(), username, t)
					if err == nil && ok {
						next(w, r)
						return
					}
				}
			}
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}
