package lpc

import "bytes"

const minAlloc = 64

type BitWriter struct {
	bit  uint8
	bitN int

	Buffer []byte
}

func NewBitWriter(size int) *BitWriter {
	if size < 8 {
		size = minAlloc
	}

	return &BitWriter{
		bit:    0,
		bitN:   7,
		Buffer: make([]byte, 0, size),
	}
}

func (b *BitWriter) WriteBits(bits ...bool) {
	if b.Buffer == nil {
		b.bit = 0
		b.bitN = 7
		b.Buffer = make([]byte, 0, minAlloc)
	}

	for i := range bits {
		b.writeBit(bits[i])
	}
}

func (b *BitWriter) Flush() {
	if b.bit == 0 && b.bitN == 7 {
		return
	}

	if b.bitN > 0 {
		b.bit >>= b.bitN + 1
	}

	b.Buffer = append(b.Buffer, b.bit)
	b.bit = 0
	b.bitN = 7
}

func (b *BitWriter) writeBit(bit bool) {
	if bit {
		b.bit += 1 << b.bitN
	}

	b.bitN--

	if b.bitN < 0 {
		b.Flush()
	}
}

type BitBuffer struct {
	bit  uint8
	bitN int

	Buffer *bytes.Buffer
}

func NewBitBuffer(size int) *BitBuffer {
	var buf []byte

	if size > 1 {
		buf = make([]byte, 0, size)
	}

	return &BitBuffer{
		bit:    0,
		bitN:   7,
		Buffer: bytes.NewBuffer(buf),
	}
}

func (b *BitBuffer) WriteBits(bits ...bool) {
	if b.Buffer == nil {
		b.bit = 0
		b.bitN = 7
		b.Buffer = bytes.NewBuffer(nil)
	}

	for i := range bits {
		b.writeBit(bits[i])
	}
}

func (b *BitBuffer) Flush() {
	if b.bit == 0 && b.bitN == 7 {
		return
	}

	if b.bitN > 0 {
		b.bit >>= b.bitN + 1
	}

	b.Buffer.WriteByte(b.bit)
	b.bit = 0
	b.bitN = 7
}

func (b *BitBuffer) writeBit(bit bool) {
	if bit {
		b.bit += 1 << b.bitN
	}

	b.bitN--

	if b.bitN < 0 {
		b.Flush()
	}
}
