package endpoints

import (
	"encoding/json"

	gojson "github.com/goccy/go-json"
)

func NewEncoder(encoderType string) EncodeDecoder {
	switch encoderType {
	case "json":
		return gojsonEnc{}
	case "stdjson":
		return stdjsonEnc{}
	default:
		return gojsonEnc{}
	}
}

type Encoder interface {
	Encode(any) ([]byte, error)
}

type Decoder interface {
	Decode([]byte, any) error
}

type EncodeDecoder interface {
	Encoder
	Decoder
}

type stdjsonEnc struct{}

func (stdjsonEnc) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}
func (stdjsonEnc) Decode(b []byte, v any) error {
	return json.Unmarshal(b, v)
}

type gojsonEnc struct{}

func (gojsonEnc) Encode(v any) ([]byte, error) {
	return gojson.Marshal(v)
}
func (gojsonEnc) Decode(b []byte, v any) error {
	return gojson.Unmarshal(b, v)
}
