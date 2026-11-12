package pcm

import (
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/zalgonoise/x/audio"
)

// Int is a linear integer PCM codec of 8, 16, 24 or 32 bits.
//
// Build it with NewInt to validate Bits; the methods of an Int with any other Bits panic, as a misconfigured codec is a
// programming error rather than bad input.
type Int struct {
	Bits int // 8, 16, 24 or 32
	// Order is the byte order of multi-byte samples; nil means little-endian. Any binary.ByteOrder works.
	Order binary.ByteOrder
	// Unsigned stores samples with a 2^(Bits-1) offset, as 8-bit WAV does (0x80 is silence).
	Unsigned bool
}

// NewInt returns an Int codec, or an error matching audio.ErrUnsupportedFormat for a width other than 8, 16, 24 or 32.
func NewInt(bits int, order binary.ByteOrder, unsigned bool) (Int, error) {
	c := Int{Bits: bits, Order: order, Unsigned: unsigned}
	if !c.valid() {
		return Int{}, fmt.Errorf("%w: pcm: %d-bit integer samples", audio.ErrUnsupportedFormat, bits)
	}

	return c, nil
}

func (c Int) valid() bool { return c.Bits == 8 || c.Bits == 16 || c.Bits == 24 || c.Bits == 32 }

func (c Int) check() {
	if !c.valid() {
		panic(fmt.Sprintf("pcm: Int.Bits must be 8, 16, 24 or 32, not %d", c.Bits))
	}
}

// SampleSize implements Codec.
func (c Int) SampleSize() int {
	c.check()

	return c.Bits / 8
}

func (c Int) offset() int64 {
	if c.Unsigned {
		return 1 << (c.Bits - 1)
	}

	return 0
}

// Decode implements Codec.
func (c Int) Decode(dst []float64, src []byte) int {
	size := c.SampleSize()
	n := min(len(dst), len(src)/size)
	big, offset := bigEndian(c.Order), c.offset()

	for i := range n {
		dst[i] = Normalize(c.load(src[i*size:i*size+size], big)-offset, c.Bits)
	}

	return n
}

// AppendEncode implements Codec.
func (c Int) AppendEncode(dst []byte, src []float64) []byte {
	size := c.SampleSize()
	big, offset := bigEndian(c.Order), c.offset()

	for _, f := range src {
		dst = store(dst, Quantize(f, c.Bits)+offset, size, big)
	}

	return dst
}

// load reads one sample as a signed integer (or an unsigned one, when the codec is unsigned).
func (c Int) load(b []byte, big bool) int64 {
	var u uint32

	if big {
		for _, x := range b {
			u = u<<8 | uint32(x)
		}
	} else {
		for _, v := range slices.Backward(b) {
			u = u<<8 | uint32(v)
		}
	}

	if c.Unsigned {
		return int64(u)
	}

	shift := 32 - c.Bits

	return int64(int32(u<<shift) >> shift)
}

func store(dst []byte, v int64, size int, big bool) []byte {
	if big {
		for i := size - 1; i >= 0; i-- {
			dst = append(dst, byte(v>>(8*i)))
		}

		return dst
	}

	for i := range size {
		dst = append(dst, byte(v>>(8*i)))
	}

	return dst
}

// bigEndian reports whether order stores the most significant byte first. It asks the order itself, so wrappers and
// binary.NativeEndian work too; nil is little-endian.
func bigEndian(order binary.ByteOrder) bool {
	if order == nil {
		return false
	}

	var b [2]byte

	order.PutUint16(b[:], 1)

	return b[1] == 1
}
