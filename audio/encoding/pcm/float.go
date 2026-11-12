package pcm

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/zalgonoise/x/audio"
)

// Float is an IEEE-754 floating point codec of 32 or 64 bits. Values are stored as-is: they are not clipped to [-1, 1].
//
// Build it with NewFloat to validate Bits; the methods of a Float with any other Bits panic, as a misconfigured codec is
// a programming error rather than bad input.
type Float struct {
	Bits int // 32 or 64
	// Order is the byte order; nil means little-endian. Any binary.ByteOrder works.
	Order binary.ByteOrder
}

// NewFloat returns a Float codec, or an error matching audio.ErrUnsupportedFormat for a width other than 32 or 64.
func NewFloat(bits int, order binary.ByteOrder) (Float, error) {
	c := Float{Bits: bits, Order: order}
	if !c.valid() {
		return Float{}, fmt.Errorf("%w: pcm: %d-bit float samples", audio.ErrUnsupportedFormat, bits)
	}

	return c, nil
}

func (c Float) valid() bool { return c.Bits == 32 || c.Bits == 64 }

// SampleSize implements Codec.
func (c Float) SampleSize() int {
	if !c.valid() {
		panic(fmt.Sprintf("pcm: Float.Bits must be 32 or 64, not %d", c.Bits))
	}

	return c.Bits / 8
}

func (c Float) order() binary.ByteOrder {
	if c.Order == nil {
		return binary.LittleEndian
	}

	return c.Order
}

// Decode implements Codec.
func (c Float) Decode(dst []float64, src []byte) int {
	size := c.SampleSize()
	n := min(len(dst), len(src)/size)
	order := c.order()

	for i := range n {
		b := src[i*size : i*size+size]

		if c.Bits == 64 {
			dst[i] = math.Float64frombits(order.Uint64(b))

			continue
		}

		dst[i] = float64(math.Float32frombits(order.Uint32(b)))
	}

	return n
}

// AppendEncode implements Codec.
func (c Float) AppendEncode(dst []byte, src []float64) []byte {
	size := c.SampleSize()
	order := c.order()

	var b [8]byte

	for _, f := range src {
		if c.Bits == 64 {
			order.PutUint64(b[:], math.Float64bits(f))
		} else {
			order.PutUint32(b[:], math.Float32bits(float32(f)))
		}

		dst = append(dst, b[:size]...)
	}

	return dst
}
