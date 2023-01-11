package ghttp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
)

// HTTPWriter writes an object to a http.ResponseWriter as a HTTP response
type HTTPWriter interface {
	// WriteHTTP writes the contents of the object to the http.ResponseWriter `w`
	WriteHTTP(ctx context.Context, w http.ResponseWriter)
}

type okResponse[T any] struct {
	Message string            `json:"message,omitempty"`
	Data    *T                `json:"data,omitempty"`
	Status  int               `json:"-"`
	Headers map[string]string `json:"-"`
}

type errResponse struct {
	Message string            `json:"message,omitempty"`
	Error   string            `json:"error,omitempty"`
	Status  int               `json:"-"`
	Headers map[string]string `json:"-"`
}

// okResponse creates a generic data structure for OK HTTP responses
func OKResponse[T any](status int, message string, data *T, headers map[string]string) *okResponse[T] {
	return &okResponse[T]{
		Message: message,
		Data:    data,
		Status:  status,
		Headers: headers,
	}
}

// errResponse creates a data structure for not-OK HTTP responses
func ErrResponse(status int, message string, err error, headers map[string]string) *errResponse {
	return &errResponse{
		Message: message,
		Error:   err.Error(),
		Status:  status,
		Headers: headers,
	}
}

// WriteHTTP writes the contents of the object to the http.ResponseWriter `w`
func (r okResponse[T]) WriteHTTP(ctx context.Context, w http.ResponseWriter) {
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

// WriteHTTP writes the contents of the object to the http.ResponseWriter `w`
func (r errResponse) WriteHTTP(ctx context.Context, w http.ResponseWriter) {
	ctx, s := spanner.Start(ctx, "http.HttpResponse.WriteHTTP")
	defer s.End()
	s.Add(
		attr.Int("http_status", r.Status),
		attr.String("with_error", r.Error),
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
