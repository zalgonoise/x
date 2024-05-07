package pages

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
)

func AsBase64JSON[T any](value T) (string, error) {
	buf := bytes.NewBuffer(nil)

	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	defer encoder.Close()

	err := json.NewEncoder(encoder).Encode(&value)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func FromBase64JSON[T any](token string) (*T, error) {
	value := new(T)

	if err := json.NewDecoder(base64.NewDecoder(base64.StdEncoding, strings.NewReader(token))).Decode(value); err != nil {
		return nil, err
	}

	return value, nil
}
