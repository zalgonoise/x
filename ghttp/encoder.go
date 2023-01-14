package ghttp

import (
	"bytes"
	"context"
	stdjson "encoding/json"
	"errors"
	"io"
	"sync"

	json "github.com/goccy/go-json"
)

var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

const bufferCap int = 4096

var (
	ErrInvalidEncoder = errors.New("invalid encoder - no Encode(any) error method")
	ErrInvalidDecoder = errors.New("invalid decoder - no Decode(any) error method")
)

// ContextResponseEncoder is a type used to identify encoders in Contexts
type ContextResponseEncoder string

// ResponseEncoderKey is the common key used by this package to store an
// EncodeDecoder in a Context
const ResponseEncoderKey ContextResponseEncoder = "encoder"

// Enc returns the EncodeDecoder from the input Context `ctx`, or creates a new
// JSON encoder if the context doesn't have one
func Enc(ctx context.Context) EncodeDecoder {
	v := ctx.Value(ResponseEncoderKey)
	if v != nil {
		if e, ok := v.(EncodeDecoder); ok {
			return e
		}
	}
	return JSON()
}

type encDec struct {
	enc func(io.Writer) any
	dec func(io.Reader) any
}

// NewEncodeDecoder creates an EncodeDecoder based off of encoder and
// decoders' init functions. As these functions return a pointer to their
// own Encoder / Decoder types, they are cast as `any`.
//
// The EncodeDecoder will expect the respective Encode/Decode(any) error
// interface to be available otherwise it will return an error when
// performing that action
func NewEncodeDecoder(
	enc func(io.Writer) any,
	dec func(io.Reader) any,
) EncodeDecoder {
	return encDec{enc, dec}
}

// Encode encodes any item into a slice of bytes, returning also an error
func (ed encDec) Encode(v any) ([]byte, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	enc := ed.enc(buf)
	if e, ok := enc.(interface{ Encode(any) error }); ok {
		err := e.Encode(v)
		if err != nil {
			return nil, err
		}
		b := buf.Bytes()
		if buf.Cap() > bufferCap {
			return b, nil
		}
		buf.Reset()
		bufferPool.Put(buf)
		return b, nil
	}
	if buf.Cap() > bufferCap {
		return nil, ErrInvalidEncoder
	}
	buf.Reset()
	bufferPool.Put(buf)
	return nil, ErrInvalidEncoder

}

// Decode converts the data in the slice of bytes into the input object, returning an error
func (ed encDec) Decode(b []byte, v any) error {
	buf := bytes.NewBuffer(b)
	dec := ed.dec(buf)
	if d, ok := dec.(interface{ Decode(any) error }); ok {
		err := d.Decode(v)
		if err != nil {
			return err
		}
		return nil
	}
	return ErrInvalidDecoder
}

// Encoder encodes any item into a slice of bytes, returning also an error
type Encoder interface {
	// Encode encodes any item into a slice of bytes, returning also an error
	Encode(any) ([]byte, error)
}

// Decoder converts the data in the slice of bytes into the input object, returning an error
type Decoder interface {
	// Decode converts the data in the slice of bytes into the input object, returning an error
	Decode([]byte, any) error
}

// EncodeDecoder combines an Encoder with a Decoder
type EncodeDecoder interface {
	// Encoder encodes any item into a slice of bytes, returning also an error
	Encoder
	// Decoder converts the data in the slice of bytes into the input object, returning an error
	Decoder
}

// MarshalUnmarshaler is a marshals and unmarshals data
type MarshalUnmarshaler interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(b []byte, v any) error
}

// New creates an EncodeDecoder from a MarshalUnmarshaler
func FromMarshaler(mu MarshalUnmarshaler) EncodeDecoder {
	return muEncDec{mu}
}

type muEncDec struct {
	mu MarshalUnmarshaler
}

// Encode encodes any item into a slice of bytes, returning also an error
func (ed muEncDec) Encode(v any) ([]byte, error) {
	return ed.mu.Marshal(v)
}

// Decode converts the data in the slice of bytes into the input object, returning an error
func (ed muEncDec) Decode(b []byte, v any) error {
	return ed.mu.Unmarshal(b, v)
}

// JSON returns a goccy/go-json EncodeDecoder
func JSON() EncodeDecoder {
	return encDec{
		enc: func(w io.Writer) any { return json.NewEncoder(w) },
		dec: func(w io.Reader) any { return json.NewDecoder(w) },
	}
}

// StdJSON returns an encoding/json std library EncodeDecoder
func StdJSON() EncodeDecoder {
	return encDec{
		enc: func(w io.Writer) any { return stdjson.NewEncoder(w) },
		dec: func(w io.Reader) any { return stdjson.NewDecoder(w) },
	}
}
