package wav

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// WavHeader describes the header of a WAV file (or buffer).
//
// The structure is defined as seen in the WAV file format, and can
// be quickly encoded / decoded into binary format as-is
//
// Reference: http://soundfile.sapp.org/doc/WaveFormat/
type WavHeader struct {
	ChunkID       [4]byte // 1-4
	ChunkSize     uint32  // 5-8
	Format        [4]byte // 9-12
	Subchunk1ID   [4]byte // 13-16
	Subchunk1Size uint32  // 17-20
	AudioFormat   uint16  // 21-22
	NumChannels   uint16  // 23-24
	SampleRate    uint32  // 25-28
	ByteRate      uint32  // 29-32
	BlockAlign    uint16  // 33-34
	BitsPerSample uint16  // 35-36
}

// HeaderFrom extracts a WAV header from an input chunk of bytes; returning a
// pointer to a WavHeader, and an error if the data is invalid
func HeaderFrom(buf []byte) (*WavHeader, error) {
	r := bytes.NewReader(buf)
	var header = new(WavHeader)

	if err := binary.Read(r, binary.LittleEndian, header); err != nil {
		return nil, err
	}

	return header, ValidateHeader(header)
}

func ValidateHeader(header *WavHeader) error {
	if string(header.ChunkID[:]) != string(defaultChunkID[:]) {
		return fmt.Errorf("%w: ChunkID %s", ErrInvalidHeader, string(header.ChunkID[:]))
	}

	if string(header.Format[:]) != string(defaultFormat[:]) {
		return fmt.Errorf("%w: Format %s", ErrInvalidHeader, string(header.Format[:]))
	}

	if _, ok := validSampleRates[header.SampleRate]; !ok {
		return fmt.Errorf("%w: SampleRate %d", ErrInvalidSampleRate, header.SampleRate)
	}

	if _, ok := validBitDepths[header.BitsPerSample]; !ok {
		return fmt.Errorf("%w: BitsPerSample %d", ErrInvalidBitDepth, header.BitsPerSample)
	}

	if _, ok := validNumChannels[header.NumChannels]; !ok {
		return fmt.Errorf("%w: NumChannels %d", ErrInvalidNumChannels, header.NumChannels)
	}

	if _, ok := validAudioFormats[header.AudioFormat]; !ok {
		return fmt.Errorf("%w: AudioFormat %d", ErrInvalidAudioFormat, header.AudioFormat)
	}

	return nil
}

// Bytes casts a WavHeader as a slice of bytes, by binary-encoding the
// object with a little-endian (LE) byte order
func (h *WavHeader) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, headerLen))
	_ = binary.Write(buf, binary.LittleEndian, h)
	return buf.Bytes()
}
