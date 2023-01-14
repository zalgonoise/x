package ghttp

import (
	"context"
	"net/http"
)

// WithEncoder adds an encoder to an HTTP request's context
func WithEncoder(e EncodeDecoder) MiddlewareFn {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ResponseEncoderKey, e)
			r = r.WithContext(ctx)
			next(w, r)
		}
	}
}
