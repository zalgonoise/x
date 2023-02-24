package wav

import (
	"errors"
)

type err string

func (e err) Error() string { return (string)(e) }

const (
	ErrInvalidNumChannels err = "invalid number of channels"
	ErrInvalidSampleRate  err = "invalid sample rate"
	ErrInvalidBitDepth    err = "invalid bit depth"
	ErrShortBuffer        err = "data buffer is too short"
	ErrZeroChunks         err = "no buffered chunks available"
)

type Wav struct {
	Header *WavHeader
	Chunks []*SubChunk
	Junk   []byte
	Data   []int
}

func New(sampleRate uint32, bitDepth, numChannels uint16) (*Wav, error) {
	var err error
	var errs []error

	if _, ok := validNumChannels[numChannels]; !ok {
		errs = append(errs, ErrInvalidNumChannels)
		numChannels = 1
	}
	if _, ok := validSampleRates[sampleRate]; !ok {
		errs = append(errs, ErrInvalidSampleRate)
		sampleRate = sampleRate44100
	}

	if _, ok := validBitDepths[bitDepth]; !ok {
		errs = append(errs, ErrInvalidBitDepth)
		bitDepth = bitDepth16
	}
	switch len(errs) {
	case 0:
		err = nil
	case 1:
		err = errs[0]
	default:
		err = errors.Join(errs...)
	}

	return &Wav{
		Header: &WavHeader{
			ChunkID:       defaultChunkID,
			ChunkSize:     0,
			Format:        defaultFormat,
			Subchunk1ID:   defaultSubchunk1ID,
			Subchunk1Size: 16,
			AudioFormat:   1,
			NumChannels:   numChannels,
			SampleRate:    sampleRate,
			ByteRate:      sampleRate * uint32(bitDepth) * uint32(numChannels) / 8,
			BlockAlign:    bitDepth * numChannels / 8,
			BitsPerSample: bitDepth,
		},
		Chunks: []*SubChunk{
			{
				Subchunk2ID:   defaultSubchunk2ID,
				Subchunk2Size: 0,
			},
		},
	}, err
}
