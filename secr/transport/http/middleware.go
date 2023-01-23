package http

import (
	"net/http"

	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/authz"
)

// WithAuth is middleware to validate JWT in request headers, for sensitive endpoints
func (s *server) WithAuth() ghttp.MiddlewareFn {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if token, ok := getToken(r); ok {
				u, err := s.s.ParseToken(r.Context(), token)
				if err == nil {
					// wrap caller info in context
					next(w, authz.SignRequest(u.Username, r))
					return
				}
			}

			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}
