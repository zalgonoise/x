package spanner

import "encoding/json"

// Encoder encodes any type into a byte slice or an error
type Encoder interface {
	Encode(any) ([]byte, error)
}

// Decoder decodes the input byte slice into the input type any, and returns
// an error
type Decoder interface {
	Decode([]byte, any) error
}

// EncodeDecoder is able to both encode and decode data
type EncodeDecoder interface {
	Encoder
	Decoder
}

type jsonEnc struct{}

// Encoder encodes any type into a byte slice or an error
func (jsonEnc) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Decoder decodes the input byte slice into the input type any, and returns
// an error
func (jsonEnc) Decode(b []byte, v any) error {
	return json.Unmarshal(b, v)
}
