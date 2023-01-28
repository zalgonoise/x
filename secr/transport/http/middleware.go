package http

import (
	"net/http"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/authz"
)

// WithAuth is middleware to validate JWT in request headers, for sensitive endpoints
func (s *server) WithAuth() ghttp.MiddlewareFn {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, span := spanner.Start(r.Context(), "http.WithAuth")
			r = r.WithContext(ctx)
			defer span.End()

			if token, ok := getToken(r); ok {
				u, err := s.s.ParseToken(ctx, token)
				if err == nil {

					span.Event("auth validated successfully")
					// wrap caller info in context
					next(w, authz.SignRequest(u.Username, r))
					return
				}
				span.Event("auth error", attr.String("error", err.Error()))
			}

			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}
