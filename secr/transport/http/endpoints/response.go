package endpoints

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
)

var (
	ErrInvalidBody   = errors.New("invalid body")
	ErrInvalidJSON   = errors.New("body contains invalid JSON")
	ErrInternal      = errors.New("internal error")
	ErrResponseWrite = errors.New("failed to write response")
)

// ContextResponseEncoder is a type used to identify encoders in Contexts
type ContextResponseEncoder string

// ResponseEncoderKey is the common key used by this package to store an
// EncodeDecoder in a Context
const ResponseEncoderKey ContextResponseEncoder = "encoder"

// enc returns the EncodeDecoder from the input Context `ctx`, or creates a new
// JSON encoder if the context doesn't have one
func enc(ctx context.Context) EncodeDecoder {
	var enc EncodeDecoder = nil

	v := ctx.Value(ResponseEncoderKey)
	if v != nil {
		if e, ok := v.(EncodeDecoder); ok {
			enc = e
		}
	}
	if enc == nil {
		return NewEncoder("json")
	}
	return enc
}

// HTTPWriter writes an object to a http.ResponseWriter as a HTTP response
type HTTPWritter interface {
	// WriteHTTP writes the contents of the object to the http.ResponseWriter `w`
	WriteHTTP(ctx context.Context, w http.ResponseWriter)
}

// httpResponse is a generic type for HTTP responses, both OK and not-OK
type HttpResponse[T any] struct {
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    *T     `json:"data,omitempty"`
	Status  int    `json:"-"`
}

// newResponse builds a httpResponse object for type T derived from the input data (*T)
//
// It takes in an int `status`, a string `message` and either an error or data.
//   - if the error is nil, it is assumed that it is an OK response and the provided *T `data`
//     is stored in the response, and the error ignored
//   - if the error is not nil, it is assumed that the response is not-OK, so the *T `data` is
//     ignored, and the error stored in the response
func NewResponse[T any](status int, message string, err error, data *T) HttpResponse[T] {
	if err != nil {
		return HttpResponse[T]{
			Success: false,
			Message: message,
			Error:   err.Error(),
			Data:    nil,
			Status:  status,
		}
	}

	return HttpResponse[T]{
		Success: true,
		Message: message,
		Error:   "",
		Data:    data,
		Status:  status,
	}
}

// WriteHTTP writes the contents of the object to the http.ResponseWriter `w`
func (r HttpResponse[T]) WriteHTTP(ctx context.Context, w http.ResponseWriter) {
	ctx, s := spanner.Start(ctx, "http.HttpResponse.WriteHTTP")
	defer s.End()
	s.Add(
		attr.Int("http_status", r.Status),
		attr.String("for_type", fmt.Sprintf("%T", *new(T))),
		attr.New("response", r),
	)

	enc := enc(ctx)

	w.WriteHeader(r.Status)
	response, err := enc.Encode(r)
	if err != nil {
		s.Event("failed to encode response", attr.String("error", err.Error()))
	}

	n, err := w.Write(response)
	if err != nil {
		s.Event("failed to write response", attr.String("error", err.Error()))
	}

	s.Event("response written successfully", attr.String("raw", string(response)), attr.Int("bytes_written", n))
}

// readBody reads the data in the Body of *http.Request `r` as a bytes buffer,
// and attempts to decode it into an object of type T by creating a new pointer of
// this type and decoding the buffer into it
//
// Returns a pointer to the object T and an error
func readBody[T any](ctx context.Context, r *http.Request) (*T, error) {
	ctx, s := spanner.Start(ctx, "http.readBody")
	defer s.End()
	s.Add(attr.String("for_type", fmt.Sprintf("%T", *new(T))))

	b, err := io.ReadAll(r.Body)
	if err != nil {
		s.Event("error reading body", attr.New("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", ErrInvalidBody, err)
	}
	item := new(T)
	err = enc(ctx).Decode(b, item)
	if err != nil {
		s.Event("error decoding buffer", attr.New("error", err.Error()), attr.String("buffer", string(b)))
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	s.Event("decoded request body", attr.Ptr("item", item))
	return item, nil
}
