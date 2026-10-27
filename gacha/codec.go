package gacha

import "encoding/json"

// Marshaler converts an item into its raw, stored form.
type Marshaler[T, R any] interface {
	Marshal(item T) (R, error)
}

// MarshalerFunc is a function type that implements Marshaler.
type MarshalerFunc[T, R any] func(item T) (R, error)

// Marshal calls fn(item).
func (fn MarshalerFunc[T, R]) Marshal(item T) (R, error) {
	return fn(item)
}

// Unmarshaler converts a raw, stored form back into an item.
type Unmarshaler[T, R any] interface {
	Unmarshal(data R) (T, error)
}

// UnmarshalerFunc is a function type that implements Unmarshaler.
type UnmarshalerFunc[T, R any] func(data R) (T, error)

// Unmarshal calls fn(data).
func (fn UnmarshalerFunc[T, R]) Unmarshal(data R) (T, error) {
	return fn(data)
}

// Codec is both a Marshaler and an Unmarshaler for the same pair of types.
type Codec[T, R any] interface {
	Marshaler[T, R]
	Unmarshaler[T, R]
}

type codec[T, R any] struct {
	Marshaler[T, R]
	Unmarshaler[T, R]
}

// NewCodec joins a Marshaler and an Unmarshaler into a Codec.
func NewCodec[T, R any](m Marshaler[T, R], u Unmarshaler[T, R]) Codec[T, R] {
	return codec[T, R]{Marshaler: m, Unmarshaler: u}
}

type jsonCodec[T any] struct{}

func (jsonCodec[T]) Marshal(item T) ([]byte, error) {
	return json.Marshal(item)
}

func (jsonCodec[T]) Unmarshal(data []byte) (T, error) {
	var item T

	err := json.Unmarshal(data, &item)

	return item, err
}

// JSON returns a Codec that encodes items as JSON, using encoding/json.
func JSON[T any]() Codec[T, []byte] {
	return jsonCodec[T]{}
}

type bytesCodec struct{}

func (bytesCodec) Marshal(item []byte) ([]byte, error) {
	return append([]byte{}, item...), nil
}

func (bytesCodec) Unmarshal(data []byte) ([]byte, error) {
	return data, nil
}

// Bytes returns a Codec for items that already are byte slices.
//
// Marshal copies the item, so the caller can reuse its buffer once the write returns, even with a repository that
// stores the raw values as given (such as the memory repository). Unmarshal returns the data as is.
func Bytes() Codec[[]byte, []byte] {
	return bytesCodec{}
}

type stringCodec struct{}

func (stringCodec) Marshal(item string) ([]byte, error) {
	return []byte(item), nil
}

func (stringCodec) Unmarshal(data []byte) (string, error) {
	return string(data), nil
}

// String returns a Codec that stores strings as their bytes.
func String() Codec[string, []byte] {
	return stringCodec{}
}

type identityCodec[T any] struct{}

func (identityCodec[T]) Marshal(item T) (T, error) {
	return item, nil
}

func (identityCodec[T]) Unmarshal(data T) (T, error) {
	return data, nil
}

// Identity returns a Codec that passes items through unchanged. Use it with a repository that is generic over the
// raw type (such as the memory repository) to store items without marshaling them, or with a poller over a
// repository view to handle the raw data directly.
func Identity[T any]() Codec[T, T] {
	return identityCodec[T]{}
}
