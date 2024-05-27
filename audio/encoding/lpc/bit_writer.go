package lpc

const minAlloc = 64

type BitWriter struct {
	bit  uint8
	bitN int

	buffer []byte
}

func NewBitWriter(size int) *BitWriter {
	if size < 8 {
		size = minAlloc
	}

	return &BitWriter{
		bit:    0,
		bitN:   7,
		buffer: make([]byte, 0, size),
	}
}

func (b *BitWriter) WriteBits(bits ...bool) {
	if b.buffer == nil {
		b.bit = 0
		b.bitN = 7
		b.buffer = make([]byte, 0, minAlloc)
	}

	for i := range bits {
		b.writeBit(bits[i])
	}
}

func (b *BitWriter) writeBit(bit bool) {
	if bit {
		b.bit += 1 << b.bitN
	}

	b.bitN--

	if b.bitN < 0 {
		b.buffer = append(b.buffer, b.bit)
		b.bit = 0
		b.bitN = 7
	}
}
