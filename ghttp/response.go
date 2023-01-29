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
	Data    *T                `json:"data,omitempty"`
	Headers map[string]string `json:"-"`
}

// NewResponse creates a generic HTTP response, initialized with a status and body message
func NewResponse[T any](status int, msg string) *Response[T] {
	return &Response[T]{
		Status:  status,
		Message: msg,
		Headers: make(map[string]string),
	}
}

// WithData is a chaining method to add a data object to a response
func (r *Response[T]) WithData(data *T) *Response[T] {
	r.Data = data
	return r
}

// WithHeaders is a chaining method to add headers to a response
func (r *Response[T]) WithHeaders(headers map[string]string) *Response[T] {
	r.Headers = headers
	return r
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
	w.Header().Set("Content-Type", "application/json")

	n, err := w.Write(response)
	if err != nil {
		s.Event("failed to write response", attr.String("error", err.Error()))
	}

	s.Event("response written successfully", attr.Int("bytes_written", n))
}
