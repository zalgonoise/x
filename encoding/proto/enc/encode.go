package enc

import (
	"bytes"

	"github.com/zalgonoise/x/conv"
)

type Person struct {
	Name    string
	Age     uint64
	ID      uint64
	IsAdmin uint64
}

// 1 byte per field
const minSize = 4

type Encoder struct {
	b *bytes.Buffer
}

type number interface {
	~uint | ~uint16 | ~uint32 | ~uint64 | ~int | ~int16 | ~int32 | ~int64
}

func byteLen[T number](v T) (size int) {
	for i := 0; i < 8; i++ {
		v = v >> 8
		if v == 0 {
			return i + i
		}
	}
	return 0
}

func (p Person) Bytes() []byte {
	// init buffer with expected min size
	e := NewEncoder(
		minSize +
			byteLen(p.Age) +
			byteLen(p.IsAdmin) +
			byteLen(p.ID) +
			byteLen(len(p.Name)) +
			len(p.Name)>>8 + 1 +
			len(p.Name))
	e.b.WriteByte(18)
	e.EncodeLengthDelimited([]byte(p.Name))

	e.b.WriteByte(24)
	e.EncodeVarint(p.Age)

	e.b.WriteByte(32)
	e.EncodeVarint(p.IsAdmin)
	e.b.WriteByte(32)
	e.EncodeVarint(p.ID)
	return e.Bytes()
}

func NewEncoder(size int) Encoder {
	if size == 0 {
		size = minSize
	}
	return Encoder{
		b: bytes.NewBuffer(make([]byte, 0, size)),
	}
}

func (w Encoder) EncodeVarint(value uint64) int {
	i := 0
	for value >= 0x80 {
		_ = w.b.WriteByte(byte(value) | 0x80)
		value >>= 7
		i++
	}
	_ = w.b.WriteByte(byte(value))
	return i + 1
}

func zigZagEncode(n int64) uint64 {
	return uint64((n << 1) ^ (n >> 63))
}

func float32Encode(n float32) []byte {
	return conv.Float32To(n)
}

func float64Encode(n float64) []byte {
	return conv.Float64To(n)
}

func boolEncode(v bool) byte {
	if v {
		return 1
	}
	return 0
}

func (w Encoder) EncodeLengthDelimited(value []byte) int {
	n := w.EncodeVarint(uint64(len(value)))
	_, _ = w.b.Write(value)
	return n + len(value)

}
func (w Encoder) EncodeField(fieldNumber int, wireType int, value []byte) int {
	n := w.EncodeVarint(uint64((fieldNumber << 3) | wireType))
	return n + w.EncodeLengthDelimited(value)
}
func (w Encoder) EncodeVarintField(fieldNumber int, value uint64) int {
	n := w.EncodeVarint(uint64((fieldNumber << 3)))
	return n + w.EncodeVarint(value)
}

func (w Encoder) String() string {
	return w.b.String()
}

func (w Encoder) Bytes() []byte {
	return w.b.Bytes()
}
