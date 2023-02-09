package enc

import "bytes"

type Person struct {
	Name    string
	Age     uint64
	ID      uint64
	IsAdmin uint64
}

// 1 byte per field, plus 1 byte for LEN field's length
const minSize = 5

type Encoder struct {
	b *bytes.Buffer
}

type number interface {
	~uint | ~uint16 | ~uint32 | ~uint64 | ~int | ~int16 | ~int32 | ~int64
}

func byteLen[T number](v T) int {
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

func (w Encoder) EncodeInt64(value int64) int {
	return w.EncodeVarint(uint64(value<<1) ^ uint64(value>>63))
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
