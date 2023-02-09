package enc

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
)

type IDAndWire struct {
	ID   int
	Wire int
	Name string
}

func HeaderGoString(fields ...IDAndWire) string {
	if len(fields) == 0 {
		return ""
	}

	sb := new(strings.Builder)
	sb.WriteString("var (\n")

	for _, f := range fields {
		if f.Name == "" {
			continue
		}
		n := []byte(f.Name)
		if n[0] > 96 {
			n[0] = n[0] - 32 // fast uppercase
		}
		b := (f.ID << 3) | f.Wire
		sb.WriteString("\theader")
		sb.WriteString(string(n))
		sb.WriteString(" uint64 = ")
		sb.WriteString(strconv.Itoa(b))
		sb.WriteByte('\n')
	}
	sb.WriteString(")\n\n")

	return sb.String()
}

const MaxVarintLen64 = 10

type Decoder struct {
	*bytes.Reader
}

func NewDecoder(buf []byte) *Decoder {
	return &Decoder{
		Reader: bytes.NewReader(buf),
	}
}

var (
	headerName    uint64 = 18 // {2, 2}
	headerAge     uint64 = 24 // {3, 0}
	headerIsAdmin uint64 = 32 // {4, 0}
	headerID      uint64 = 40 // {5, 0}
)

func (d *Decoder) Decode() (Person, error) {
	p, err := decodePerson(d.Reader)
	if err != nil && !errors.Is(err, io.EOF) {
		return p, err
	}
	return p, nil
}

func decodePerson(r io.Reader) (Person, error) {
	var p = Person{}
	for {
		v, err := decodeVarint(r)
		if err != nil {
			return p, err
		}
		switch v {
		case headerName:
			name, err := decodeString(r)
			if err != nil {
				return p, err
			}
			p.Name = name
		case headerAge:
			age, err := decodeVarint(r)
			if err != nil {
				return p, err
			}
			p.Age = age
		case headerIsAdmin:
			isAdmin, err := decodeVarint(r)
			if err != nil {
				return p, err
			}
			p.IsAdmin = isAdmin
		case headerID:
			id, err := decodeVarint(r)
			if err != nil {
				return p, err
			}
			p.ID = id
		case 0:
			return p, nil
		default:
			return p, errors.New("invalid header")
		}
	}
}

func decodeString(r io.Reader) (string, error) {
	length, err := decodeVarint(r)
	if err != nil {
		return "", err
	}
	data := make([]byte, length)
	_, err = r.Read(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
func decodeVarint(r io.Reader) (uint64, error) {
	var x uint64
	var s uint
	var i int
	for {
		byt := make([]byte, 1)
		_, err := r.Read(byt)
		if err != nil {
			return x, err
		}
		i++
		if i == MaxVarintLen64 {
			return 0, errors.New("varint overflow") // overflow
		}
		if byt[0] < 0x80 {
			if i == MaxVarintLen64-1 && byt[0] > 1 {
				return 0, errors.New("varint overflow") // overflow
			}
			return x | uint64(byt[0])<<s, nil
		}
		x |= uint64(byt[0]&0x7f) << s
		s += 7
	}
}

func zigZagDecode(n uint64) int64 {
	return int64((n >> 1) ^ -(n & 1))
}
