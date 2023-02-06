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

type field[T any] struct {
	num      int
	wireType int
	len      int
	value    T
}

func newField[T any](num, wtype, len int, value T) field[T] {
	return field[T]{
		num:      num,
		wireType: wtype,
		len:      len,
		value:    value,
	}
}

type Field interface {
	Value() any
	Type() int
	Num() int
	Len() int
}

func (f field[T]) Value() any {
	return f.value
}

func (f field[T]) Type() int {
	return f.wireType
}

func (f field[T]) Num() int {
	return f.num
}

func (f field[T]) Len() int {
	return f.len
}

func NewDecoder(buf []byte) *Decoder {
	return &Decoder{
		Reader: bytes.NewReader(buf),
		len:    len(buf),
	}
}

func (d *Decoder) Decode() (out map[int]Field, err error) {
	out = make(map[int]Field)

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
			out[f.Num()] = f
		case 2:
			out[f.Num()] = f
		}
	}
}

func decodeField(r io.Reader) (int, Field, error) {
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
		f := newField(num, wireType, n, v)
		return wireType, f, nil
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
		f := newField(num, wireType, n, data)
		return wireType, f, nil
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

func decodeVarint(r io.Reader) (uint64, int, error) {
	var (
		n     int
		value uint64
	)
	for shift := uint(0); ; shift += 7 {
		byt := make([]byte, 1)
		_, err := r.Read(byt)
		if err != nil {
			return value, n, err
		}
		n++
		value |= (uint64(byt[0]) & 0x7F) << shift

		if (byt[0] & 0x80) != 0 {
			continue
		}
		break
	}
	return value, n, nil
}
