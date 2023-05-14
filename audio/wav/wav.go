package wav

import (
	"bytes"
	"time"

	"github.com/zalgonoise/x/audio/osc"
)

// Wav describes the structure of WAV-encoded audio data, containing
// the WAV header (the audio metadata), a list of data.Chunk representing
// WAV subchunks (allocated usually to "data", or the PCM audio buffer, but
// also "junk"), and also a Data reference that is used as a pointer to the
// currently active (PCM data) chunk
type Wav struct {
	Header *WavHeader
	Chunks []Chunk
	Data   Chunk

	readOnly bool
	buf      *bytes.Buffer
}

// New creates a new Wav, configured with the input sample rate `sampleRate`
// (44100, 48000, etc), bit depth `bitDepth` (8, 16, 24 or 32), and a number of
// channels `numChannels` (either 1 or 2, for mono and stereo).
//
// This call returns a pointer to a Wav, and an error which is raised if the input
// data is invalid or unsupported.
//
// The returned Wav object will have its header set in every field except for
// `ChunkSize`, and both the `Wav.Chunks` and `Wav.Data` elements set to a blank data chunk
func New(sampleRate uint32, bitDepth, numChannels uint16) (*Wav, error) {
	blankData := NewChunk(bitDepth, nil)
	if blankData == nil {
		return nil, ErrInvalidBitDepth
	}

	w := &Wav{
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
		Chunks: []Chunk{blankData},
		Data:   blankData,
	}

	if err := ValidateHeader(w.Header); err != nil {
		return nil, err
	}

	return w, nil
}

// Generate wraps a call to w.Data.Generate, by passing the same sample rate
// value as configured in w.header.SampleRate
func (w *Wav) Generate(waveType osc.Type, freq int, dur time.Duration) {
	w.Data.Generate(waveType, freq, int(w.Header.SampleRate), dur)
}
