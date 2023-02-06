package types

import (
	"bytes"
	"errors"
	"io"
)

type Decoder struct {
	*bytes.Reader
	len int
}

type Field[T any] struct {
	Num      int
	WireType int
	Len      int
	Value    T
}

type GField struct {
	u *Field[uint64]
	b *Field[[]byte]
}

func (f GField) To(to any) error {
	if to == nil {
		return errors.New("pointer is not initialized")
	}
	switch t := (to).(type) {
	case *Field[uint64]:
		if f.u == nil {
			return errors.New("uint64 field is empty")
		}
		*t = *f.u
		return nil
	case *Field[[]byte]:
		if f.b == nil {
			return errors.New("uint64 field is empty")
		}
		*t = *f.b
		return nil
	}
	return errors.New("invalid type")
}

func NewField[T any](num, wtype, len int, value T) *Field[T] {
	return &Field[T]{
		Num:      num,
		WireType: wtype,
		Len:      len,
		Value:    value,
	}
}

func WrapUint64(f *Field[uint64]) *GField {
	return &GField{
		u: f,
	}
}
func WrapBytes(f *Field[[]byte]) *GField {
	return &GField{
		b: f,
	}
}

func NewDecoder(buf []byte) *Decoder {
	return &Decoder{
		Reader: bytes.NewReader(buf),
		len:    len(buf),
	}
}

func (d *Decoder) Decode() (out map[int]any, err error) {
	out = make(map[int]any)

	for {
		wt, f, err := decodeField(d.Reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return out, nil
			}
			return out, err
		}
		switch wt {
		case 0:
			out[f.u.Num] = f
		case 2:
			out[f.b.Num] = f
		}
	}
}

func decodeField(r io.Reader) (int, *GField, error) {
	num, wireType, _, err := decodeFieldHeader(r)
	if err != nil {
		return -1, nil, err
	}
	switch wireType {
	case 0:
		v, n, err := decodeVarint(r)
		if err != nil {
			return -1, nil, err
		}
		f := NewField(num, wireType, n, v)
		return wireType, WrapUint64(f), nil
	case 1:
	case 2:
		length, _, err := decodeVarint(r)
		if err != nil {
			return -1, nil, err
		}
		data := make([]byte, length)
		n, err := r.Read(data)
		if err != nil {
			return -1, nil, err
		}
		f := NewField(num, wireType, n, data)
		return wireType, WrapBytes(f), nil
	case 3:
	case 4:
	case 5:
	}
	return -1, nil, errors.New("invalid wire type")
}
func decodeFieldHeader(r io.Reader) (field, wireType, n int, err error) {
	fieldAndWire, n, err := decodeVarint(r)
	if err != nil {
		return 0, 0, 0, err
	}
	field = int(fieldAndWire >> 3)
	wireType = int(fieldAndWire & 0x7)
	return field, wireType, n, nil
}

func decodeVarint(r io.Reader) (value uint64, n int, err error) {
	for n = 0; ; n++ {
		byt := make([]byte, 1)
		n, err := r.Read(byt)
		if err != nil {
			return value, n, err
		}
		value |= (uint64(byt[0]) & 0x7F)

		if (byt[0] & 0x80) == 0 {
			break
		}
	}
	return value, n, nil
}
