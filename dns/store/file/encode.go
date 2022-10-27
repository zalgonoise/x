package file

import (
	"encoding/json"

	"gopkg.in/yaml.v2"
)

func NewEncoder(encoderType string) EncodeDecoder {
	switch encoderType {
	case "json":
		return jsonEnc{}
	case "yaml":
		return yamlEnc{}
	default:
		return yamlEnc{}
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

type jsonEnc struct{}

func (jsonEnc) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}
func (jsonEnc) Decode(b []byte, v any) error {
	return json.Unmarshal(b, v)
}

type yamlEnc struct{}

func (yamlEnc) Encode(v any) ([]byte, error) {
	return yaml.Marshal(v)
}
func (yamlEnc) Decode(b []byte, v any) error {
	return yaml.Unmarshal(b, v)
}
