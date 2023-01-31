package ghttp

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
	"github.com/zalgonoise/x/errors"
)

var (
	// ErrInvalidBody is an error that is raised when reading the HTTP request body
	ErrInvalidBody = errors.New("invalid body")
	// ErrInvalidJSON is an error that is raised when parsing the HTTP request as JSON
	ErrInvalidJSON = errors.New("body contains invalid JSON")
)

// validate ReadBody as a ParseFn
var _ ParseFn[any] = ReadBody[any]

// ReadBody reads the data in the Body of *http.Request `r` as a bytes buffer,
// and attempts to decode it into an object of type T by creating a new pointer of
// this type and decoding the buffer into it
//
// Returns a pointer to the object T and an error
func ReadBody[T any](ctx context.Context, r *http.Request) (*T, error) {
	ctx, s := spanner.Start(ctx, "ghttp.ReadBody")
	defer s.End()
	s.Add(attr.String("for_type", fmt.Sprintf("%T", *new(T))))

	b, err := io.ReadAll(r.Body)
	if err != nil {
		s.Event("error reading body", attr.New("error", err.Error()))
		return nil, errors.Join(ErrInvalidBody, err)
	}
	item := new(T)
	err = Enc(ctx).Decode(b, item)
	if err != nil {
		s.Event("error decoding buffer", attr.New("error", err.Error()), attr.String("buffer", string(b)))
		return nil, errors.Join(ErrInvalidJSON, err)
	}
	s.Event("decoded request body")
	return item, nil
}
