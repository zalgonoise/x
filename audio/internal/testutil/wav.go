package testutil

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// WAVE format tags.
const (
	FormatPCM        uint16 = 0x0001
	FormatFloat      uint16 = 0x0003
	FormatALaw       uint16 = 0x0006
	FormatMuLaw      uint16 = 0x0007
	FormatExtensible uint16 = 0xFFFE
)

// RIFF chunk IDs used by the builder and the reference decoder.
const (
	chunkFmt  = "fmt "
	chunkData = "data"
)

// Chunk is an arbitrary RIFF chunk placed around the fmt and data chunks.
type Chunk struct {
	ID   string // four characters
	Data []byte
}

// WAV describes a RIFF/WAVE file for the builder. The zero value of each override field produces a canonical,
// spec-correct file; set the overrides to craft edge cases.
type WAV struct {
	Format        uint16 // wFormatTag; 0 means FormatPCM
	Channels      uint16
	SampleRate    uint32
	BitsPerSample uint16

	// WAVE_FORMAT_EXTENSIBLE only.
	ValidBits   uint16 // 0 means BitsPerSample
	ChannelMask uint32
	SubFormat   uint16 // FormatPCM or FormatFloat; 0 means FormatPCM

	FmtSize  uint32  // 0 means 16 (PCM), 18 (float, A-law, µ-law) or 40 (extensible); below 16 truncates the chunk
	Before   []Chunk // chunks written after fmt and before data
	After    []Chunk // chunks written after data
	DataSize *uint32 // overrides the data chunk's size field (e.g. 0 or 0xFFFFFFFF for live streams)
	RIFFSize *uint32 // overrides the RIFF chunk's size field
	NoPad    bool    // omit RIFF pad bytes after odd-sized chunks
}

// EncodeInts builds a file from per-channel signed integer samples in the range of BitsPerSample.
// 8-bit samples are written unsigned (value+128), as the WAVE spec requires.
func (w WAV) EncodeInts(perChannel [][]int32) []byte {
	data := make([]byte, 0, len(perChannel)*len(first(perChannel))*int(w.BitsPerSample/8))

	for i := range len(first(perChannel)) {
		for c := range perChannel {
			data = appendInt(data, perChannel[c][i], w.BitsPerSample)
		}
	}

	return w.Encode(data)
}

// EncodeFloats builds an IEEE-float file (32 or 64 bits) from per-channel samples.
func (w WAV) EncodeFloats(perChannel [][]float64) []byte {
	data := make([]byte, 0, len(perChannel)*len(first(perChannel))*int(w.BitsPerSample/8))

	for i := range len(first(perChannel)) {
		for c := range perChannel {
			if w.BitsPerSample == 64 {
				data = binary.LittleEndian.AppendUint64(data, math.Float64bits(perChannel[c][i]))

				continue
			}

			data = binary.LittleEndian.AppendUint32(data, math.Float32bits(float32(perChannel[c][i])))
		}
	}

	return w.Encode(data)
}

// Encode builds a file around an already-encoded data payload.
func (w WAV) Encode(data []byte) []byte {
	format := w.Format
	if format == 0 {
		format = FormatPCM
	}

	// computed in 32 bits: 65535 channels of 8 bits still fit in the 16-bit field
	blockAlign := uint16(uint32(w.Channels) * uint32(w.BitsPerSample) / 8)

	fmtBody := binary.LittleEndian.AppendUint16(nil, format)
	fmtBody = binary.LittleEndian.AppendUint16(fmtBody, w.Channels)
	fmtBody = binary.LittleEndian.AppendUint32(fmtBody, w.SampleRate)
	fmtBody = binary.LittleEndian.AppendUint32(fmtBody, w.SampleRate*uint32(blockAlign))
	fmtBody = binary.LittleEndian.AppendUint16(fmtBody, blockAlign)
	fmtBody = binary.LittleEndian.AppendUint16(fmtBody, w.BitsPerSample)

	fmtSize := w.FmtSize
	if fmtSize == 0 {
		switch format {
		case FormatExtensible:
			fmtSize = 40
		case FormatPCM:
			fmtSize = 16
		default:
			fmtSize = 18
		}
	}

	switch {
	case format == FormatExtensible && fmtSize >= 40:
		valid, sub := w.ValidBits, w.SubFormat
		if valid == 0 {
			valid = w.BitsPerSample
		}

		if sub == 0 {
			sub = FormatPCM
		}

		fmtBody = binary.LittleEndian.AppendUint16(fmtBody, 22) // cbSize
		fmtBody = binary.LittleEndian.AppendUint16(fmtBody, valid)
		fmtBody = binary.LittleEndian.AppendUint32(fmtBody, w.ChannelMask)
		// KSDATAFORMAT_SUBTYPE_*: {0000XXXX-0000-0010-8000-00AA00389B71}
		fmtBody = binary.LittleEndian.AppendUint16(fmtBody, sub)
		fmtBody = append(fmtBody, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x80, 0x00, 0x00, 0xAA, 0x00, 0x38, 0x9B, 0x71)
	case fmtSize >= 18:
		fmtBody = binary.LittleEndian.AppendUint16(fmtBody, uint16(fmtSize-18))
	}

	for uint32(len(fmtBody)) < fmtSize {
		fmtBody = append(fmtBody, 0)
	}

	// a FmtSize below 16 crafts a truncated (malformed) fmt chunk
	fmtBody = fmtBody[:fmtSize]

	body := []byte("WAVE")
	body = w.appendChunk(body, chunkFmt, fmtBody, nil)

	for _, c := range w.Before {
		body = w.appendChunk(body, c.ID, c.Data, nil)
	}

	body = w.appendChunk(body, chunkData, data, w.DataSize)

	for _, c := range w.After {
		body = w.appendChunk(body, c.ID, c.Data, nil)
	}

	riffSize := uint32(len(body))
	if w.RIFFSize != nil {
		riffSize = *w.RIFFSize
	}

	out := append([]byte("RIFF"), binary.LittleEndian.AppendUint32(nil, riffSize)...)

	return append(out, body...)
}

func (w WAV) appendChunk(dst []byte, id string, data []byte, sizeOverride *uint32) []byte {
	size := uint32(len(data))
	if sizeOverride != nil {
		size = *sizeOverride
	}

	dst = append(dst, id[:4]...)
	dst = binary.LittleEndian.AppendUint32(dst, size)
	dst = append(dst, data...)

	if len(data)%2 == 1 && !w.NoPad {
		dst = append(dst, 0)
	}

	return dst
}

func appendInt(dst []byte, v int32, bits uint16) []byte {
	switch bits {
	case 8:
		return append(dst, byte(v+128))
	case 16:
		return binary.LittleEndian.AppendUint16(dst, uint16(int16(v)))
	case 24:
		return append(dst, byte(v), byte(v>>8), byte(v>>16))
	default:
		return binary.LittleEndian.AppendUint32(dst, uint32(v))
	}
}

func first[T any](s [][]T) []T {
	if len(s) == 0 {
		return nil
	}

	return s[0]
}

// RefFile is the result of DecodeRef.
type RefFile struct {
	Format        uint16 // wFormatTag (or the SubFormat of an extensible file)
	Extensible    bool
	Channels      uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	ValidBits     uint16
	ChannelMask   uint32
	ChunkIDs      []string // every chunk ID in file order, after "WAVE"
	DataSize      uint32   // the data chunk's declared size
	Frames        int      // whole frames decoded from the data chunk

	Ints   [][]int32   // per-channel signed samples, for integer PCM (8-bit is converted from unsigned)
	Floats [][]float64 // per-channel samples, for IEEE float
}

// Errors returned by DecodeRef.
var (
	ErrRefNotRIFF   = errors.New("testutil: not a RIFF/WAVE file")
	ErrRefNoFmt     = errors.New("testutil: data chunk before fmt chunk")
	ErrRefNoData    = errors.New("testutil: no data chunk")
	ErrRefTruncated = errors.New("testutil: truncated chunk header")
	ErrRefFormat    = errors.New("testutil: unsupported sample format")
)

// DecodeRef is a small reference RIFF/WAVE decoder: it walks every chunk honouring sizes and pad bytes, parses fmt
// (16, 18 or 40 bytes, including WAVE_FORMAT_EXTENSIBLE) and decodes the data chunk, bounded by its declared size and by
// the end of the input. A data size of 0xFFFFFFFF (live streams) means "until the end of the input", and so does 0
// unless the RIFF size is definite (not 0 or 0xFFFFFFFF) and extends past the data chunk header, declaring chunks
// after it, in which case the data is empty. A trailing partial frame is ignored. A second fmt chunk before the data
// chunk is an error; after the first data chunk, chunks are only listed in ChunkIDs.
func DecodeRef(b []byte) (RefFile, error) {
	var f RefFile

	if len(b) < 12 || string(b[0:4]) != "RIFF" || string(b[8:12]) != "WAVE" {
		return f, ErrRefNotRIFF
	}

	var (
		haveFmt bool
		data    []byte
		found   bool
	)

	for off := 12; off < len(b); {
		if off+8 > len(b) {
			return f, ErrRefTruncated
		}

		id := string(b[off : off+4])
		size := binary.LittleEndian.Uint32(b[off+4 : off+8])
		start := off + 8
		end := start + int(size)

		if id == chunkData && (size == math.MaxUint32 || (size == 0 && !riffDeclaresMore(b, start)) || end > len(b)) {
			end = len(b)
		}

		end = min(end, len(b))

		f.ChunkIDs = append(f.ChunkIDs, id)

		switch id {
		case chunkFmt:
			if found {
				break
			}

			if haveFmt {
				return f, fmt.Errorf("%w: second fmt chunk", ErrRefFormat)
			}

			if err := f.parseFmt(b[start:end]); err != nil {
				return f, err
			}

			haveFmt = true
		case chunkData:
			if found {
				break
			}

			if !haveFmt {
				return f, ErrRefNoFmt
			}

			f.DataSize = size
			data, found = b[start:end], true
		}

		off = end
		if (end-start)%2 == 1 {
			off++ // pad byte
		}
	}

	if !found {
		return f, ErrRefNoData
	}

	err := f.decodeData(data)

	return f, err
}

// riffDeclaresMore reports whether the RIFF size of b is definite and extends past offset at.
func riffDeclaresMore(b []byte, at int) bool {
	size := binary.LittleEndian.Uint32(b[4:8])

	return size != 0 && size != math.MaxUint32 && at-8 < int(size)
}

func (f *RefFile) parseFmt(b []byte) error {
	if len(b) < 16 {
		return fmt.Errorf("%w: fmt chunk of %d bytes", ErrRefFormat, len(b))
	}

	f.Format = binary.LittleEndian.Uint16(b[0:2])
	f.Channels = binary.LittleEndian.Uint16(b[2:4])
	f.SampleRate = binary.LittleEndian.Uint32(b[4:8])
	f.ByteRate = binary.LittleEndian.Uint32(b[8:12])
	f.BlockAlign = binary.LittleEndian.Uint16(b[12:14])
	f.BitsPerSample = binary.LittleEndian.Uint16(b[14:16])
	f.ValidBits = f.BitsPerSample

	if f.Format == FormatExtensible && len(b) >= 40 {
		f.Extensible = true
		f.ValidBits = binary.LittleEndian.Uint16(b[18:20])
		f.ChannelMask = binary.LittleEndian.Uint32(b[20:24])
		f.Format = binary.LittleEndian.Uint16(b[24:26])
	}

	if f.Channels == 0 || f.BlockAlign == 0 || uint32(f.BlockAlign) != uint32(f.Channels)*uint32(f.BitsPerSample)/8 {
		return fmt.Errorf("%w: channels=%d blockAlign=%d bits=%d", ErrRefFormat, f.Channels, f.BlockAlign, f.BitsPerSample)
	}

	return nil
}

func (f *RefFile) decodeData(data []byte) error {
	f.Frames = len(data) / int(f.BlockAlign)
	width := int(f.BitsPerSample / 8)

	switch f.Format {
	case FormatPCM:
		f.Ints = make([][]int32, f.Channels)
	case FormatFloat:
		f.Floats = make([][]float64, f.Channels)
	default:
		return fmt.Errorf("%w: format tag %#04x", ErrRefFormat, f.Format)
	}

	for i := range f.Frames {
		for c := range int(f.Channels) {
			s := data[i*int(f.BlockAlign)+c*width:][:width]

			if f.Format == FormatFloat {
				switch width {
				case 4:
					f.Floats[c] = append(f.Floats[c], float64(math.Float32frombits(binary.LittleEndian.Uint32(s))))
				case 8:
					f.Floats[c] = append(f.Floats[c], math.Float64frombits(binary.LittleEndian.Uint64(s)))
				default:
					return fmt.Errorf("%w: %d-bit float", ErrRefFormat, f.BitsPerSample)
				}

				continue
			}

			switch width {
			case 1:
				f.Ints[c] = append(f.Ints[c], int32(s[0])-128)
			case 2:
				f.Ints[c] = append(f.Ints[c], int32(int16(binary.LittleEndian.Uint16(s))))
			case 3:
				f.Ints[c] = append(f.Ints[c], int32(uint32(s[0])|uint32(s[1])<<8|uint32(s[2])<<16)<<8>>8)
			case 4:
				f.Ints[c] = append(f.Ints[c], int32(binary.LittleEndian.Uint32(s)))
			default:
				return fmt.Errorf("%w: %d-bit PCM", ErrRefFormat, f.BitsPerSample)
			}
		}
	}

	return nil
}
