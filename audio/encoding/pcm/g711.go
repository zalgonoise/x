package pcm

// ITU-T G.711 companding (A-law and µ-law), after the public-domain Sun Microsystems reference (g711.c). Samples are
// 8-bit codes that expand to 16-bit linear values, normalised by 2^15.

const (
	g711Bits = 16

	muBias = 0x84
	muClip = 32635
)

//nolint:gochecknoglobals // immutable lookup tables, filled once at init
var (
	aLawTable  [256]int16
	muLawTable [256]int16
)

//nolint:gochecknoinits // builds the immutable decode tables
func init() {
	for i := range 256 {
		aLawTable[i] = aLawToLinear(byte(i))
		muLawTable[i] = muLawToLinear(byte(i))
	}
}

// ALaw is the G.711 A-law codec.
type ALaw struct{}

// SampleSize implements Codec.
func (ALaw) SampleSize() int { return 1 }

// Decode implements Codec.
func (ALaw) Decode(dst []float64, src []byte) int {
	n := min(len(dst), len(src))

	for i := range n {
		dst[i] = Normalize(int64(aLawTable[src[i]]), g711Bits)
	}

	return n
}

// AppendEncode implements Codec.
func (ALaw) AppendEncode(dst []byte, src []float64) []byte {
	for _, f := range src {
		dst = append(dst, linearToALaw(int32(Quantize(f, g711Bits))))
	}

	return dst
}

// MuLaw is the G.711 µ-law codec.
type MuLaw struct{}

// SampleSize implements Codec.
func (MuLaw) SampleSize() int { return 1 }

// Decode implements Codec.
func (MuLaw) Decode(dst []float64, src []byte) int {
	n := min(len(dst), len(src))

	for i := range n {
		dst[i] = Normalize(int64(muLawTable[src[i]]), g711Bits)
	}

	return n
}

// AppendEncode implements Codec.
func (MuLaw) AppendEncode(dst []byte, src []float64) []byte {
	for _, f := range src {
		dst = append(dst, linearToMuLaw(int32(Quantize(f, g711Bits))))
	}

	return dst
}

func aLawToLinear(a byte) int16 {
	a ^= 0x55

	t := int32(a&0x0F) << 4
	seg := (a & 0x70) >> 4

	switch seg {
	case 0:
		t += 8
	case 1:
		t += 0x108
	default:
		t += 0x108
		t <<= seg - 1
	}

	if a&0x80 != 0 {
		return int16(t)
	}

	return int16(-t)
}

func linearToALaw(pcm int32) byte {
	var mask byte = 0xD5

	pcm >>= 3 // 13-bit magnitude

	if pcm < 0 {
		mask = 0x55
		pcm = -pcm - 1
	}

	seg := byte(0)

	for end := int32(0x1F); seg < 8 && pcm > end; end = end<<1 | 1 {
		seg++
	}

	if seg >= 8 {
		return 0x7F ^ mask
	}

	aval := seg << 4

	if seg < 2 {
		aval |= byte(pcm>>1) & 0x0F
	} else {
		aval |= byte(pcm>>seg) & 0x0F
	}

	return aval ^ mask
}

func muLawToLinear(u byte) int16 {
	u = ^u

	t := (int32(u&0x0F)<<3 + muBias) << ((u & 0x70) >> 4)
	t -= muBias

	if u&0x80 != 0 {
		return int16(-t)
	}

	return int16(t)
}

func linearToMuLaw(pcm int32) byte {
	var sign byte

	if pcm < 0 {
		sign = 0x80
		pcm = -pcm
	}

	pcm = min(pcm, muClip) + muBias

	exponent := byte(7)
	for mask := int32(0x4000); pcm&mask == 0 && exponent > 0; mask >>= 1 {
		exponent--
	}

	mantissa := byte(pcm>>(exponent+3)) & 0x0F

	return ^(sign | exponent<<4 | mantissa)
}
