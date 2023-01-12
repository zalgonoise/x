package ghttp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
)

// Responder writes an object to a http.ResponseWriter as a HTTP response
type Responder interface {
	// WriteHTTP writes the contents of the object to the http.ResponseWriter `w`
	WriteHTTP(ctx context.Context, w http.ResponseWriter)
}

type Response[T any] struct {
	Status  int               `json:"-"`
	Message string            `json:"message,omitempty"`
	Error   string            `json:"error,omitempty"`
	Data    *T                `json:"data,omitempty"`
	Headers map[string]string `json:"-"`
}

func NewResponse[T any](status int, msg string) *Response[T] {
	return &Response[T]{
		Status:  status,
		Message: msg,
		Headers: make(map[string]string),
	}
}

// WriteHTTP writes the contents of the object to the http.ResponseWriter `w`
func (r *Response[T]) WriteHTTP(ctx context.Context, w http.ResponseWriter) {
	ctx, s := spanner.Start(ctx, "http.HttpResponse.WriteHTTP")
	defer s.End()
	s.Add(
		attr.Int("http_status", r.Status),
		attr.String("for_type", fmt.Sprintf("%T", *new(T))),
	)

	enc := Enc(ctx)

	w.WriteHeader(r.Status)
	response, err := enc.Encode(r)
	if err != nil {
		s.Event("failed to encode response", attr.String("error", err.Error()))
	}

	for k, v := range r.Headers {
		w.Header().Set(k, v)
	}

	n, err := w.Write(response)
	if err != nil {
		s.Event("failed to write response", attr.String("error", err.Error()))
	}

	s.Event("response written successfully", attr.Int("bytes_written", n))
}
