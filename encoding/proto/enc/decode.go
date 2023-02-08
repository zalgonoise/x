package enc

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

const MaxVarintLen64 = 10

type Decoder struct {
	*bytes.Reader
}

func NewDecoder(buf []byte) *Decoder {
	return &Decoder{
		Reader: bytes.NewReader(buf),
	}
}

type Person struct {
	Name    string
	Age     uint64
	ID      uint64
	IsAdmin uint64
}

func (d *Decoder) Decode() (Person, error) {
	p, err := decodePerson(d.Reader)
	if err != nil && !errors.Is(err, io.EOF) {
		return p, err
	}
	return p, nil
}

func decodePerson(r io.Reader) (Person, error) {
	var p Person
	for {
		num, wireType, _, err := decodeFieldHeader(r)
		if err != nil {
			return p, err
		}
		if num == 0 {
			return p, nil
		}
		switch wireType {
		case 0:
			v, _, err := decodeVarint(r)
			if err != nil {
				return p, err
			}
			switch num {
			case 3:
				p.Age = v
			case 4:
				p.IsAdmin = v
			case 5:
				p.ID = v
			default:
				return p, fmt.Errorf("invalid field number: %d", num)
			}
		case 2:
			length, _, err := decodeVarint(r)
			if err != nil {
				return p, err
			}
			data := make([]byte, length)
			_, err = r.Read(data)
			if err != nil {
				return p, err
			}
			switch num {
			case 2:
				p.Name = string(data)
			default:
				return p, errors.New("invalid field number")
			}
		}
	}
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
	var x uint64
	var s uint
	var i int
	for {
		byt := make([]byte, 1)
		_, err := r.Read(byt)
		if err != nil {
			return x, i, err
		}
		i++
		if i == MaxVarintLen64 {
			return 0, -(i + 1), errors.New("varint overflow") // overflow
		}
		if byt[0] < 0x80 {
			if i == MaxVarintLen64-1 && byt[0] > 1 {
				return 0, -(i + 1), errors.New("varint overflow") // overflow
			}
			return x | uint64(byt[0])<<s, i + 1, nil
		}
		x |= uint64(byt[0]&0x7f) << s
		s += 7
	}
}
