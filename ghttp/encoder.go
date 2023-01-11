package ghttp

import (
	"context"
	stdjson "encoding/json"

	json "github.com/goccy/go-json"
)

// ContextResponseEncoder is a type used to identify encoders in Contexts
type ContextResponseEncoder string

// ResponseEncoderKey is the common key used by this package to store an
// EncodeDecoder in a Context
const ResponseEncoderKey ContextResponseEncoder = "encoder"

// Enc returns the EncodeDecoder from the input Context `ctx`, or creates a new
// JSON encoder if the context doesn't have one
func Enc(ctx context.Context) EncodeDecoder {
	var enc EncodeDecoder = nil

	v := ctx.Value(ResponseEncoderKey)
	if v != nil {
		if e, ok := v.(EncodeDecoder); ok {
			enc = e
		}
	}
	if enc == nil {
		return JSON()
	}
	return enc
}

// Encoder encodes any item into a slice of bytes, returning also an error
type Encoder interface {
	Encode(any) ([]byte, error)
}

// Decoder converts the data in the slice of bytes into the input object, returning an error
type Decoder interface {
	Decode([]byte, any) error
}

// EncodeDecoder combines an Encoder with a Decoder
type EncodeDecoder interface {
	// Encoder encodes any item into a slice of bytes, returning also an error
	Encoder
	// Decoder converts the data in the slice of bytes into the input object, returning an error
	Decoder
}

// RawEncodeDecoder is a MarshalUnmarshaler
type RawEncodeDecoder interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(b []byte, v any) error
}

// New creates an EncodeDecoder from a RawEncodeDecoder
func New(red RawEncodeDecoder) EncodeDecoder {
	return encDec{red}
}

type encDec struct {
	e RawEncodeDecoder
}

// Encoder encodes any item into a slice of bytes, returning also an error
func (e encDec) Encode(v any) ([]byte, error) {
	return e.e.Marshal(v)
}

// Decoder converts the data in the slice of bytes into the input object, returning an error
func (e encDec) Decode(b []byte, v any) error {
	return e.e.Unmarshal(b, v)
}

func JSON() EncodeDecoder {
	return jsonEnc{}
}

type jsonEnc struct{}

// Encoder encodes any item into a slice of bytes, returning also an error
func (e jsonEnc) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Decoder converts the data in the slice of bytes into the input object, returning an error
func (e jsonEnc) Decode(b []byte, v any) error {
	return json.Unmarshal(b, v)
}

func StdJSON() EncodeDecoder {
	return stdJsonEnc{}
}

type stdJsonEnc struct{}

// Encoder encodes any item into a slice of bytes, returning also an error
func (e stdJsonEnc) Encode(v any) ([]byte, error) {
	return stdjson.Marshal(v)
}

// Decoder converts the data in the slice of bytes into the input object, returning an error
func (e stdJsonEnc) Decode(b []byte, v any) error {
	return stdjson.Unmarshal(b, v)
}
