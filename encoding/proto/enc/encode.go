package enc

import "bytes"

const minSize = 512

type Encoder struct {
	b *bytes.Buffer
}

func NewEncoder() Encoder {
	return Encoder{
		b: bytes.NewBuffer(make([]byte, 0, minSize)),
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
