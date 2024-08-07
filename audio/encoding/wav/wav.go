package wav

import (
	"bytes"
	"io"
	"sync/atomic"
	"time"

	"github.com/zalgonoise/x/audio/osc"
)

// Wav describes the structure of WAV-encoded audio data, containing
// the WAV header (the audio metadata), a list of data.Chunk representing
// WAV subchunks (allocated usually to "data", or the PCM audio buffer, but
// also "junk"), and also a Data reference that is used as a pointer to the
// currently active (PCM data) chunk.
type Wav struct {
	Header *Header
	Chunks []Chunk
	Data   Chunk

	readOnly atomic.Bool
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
// `ChunkSize`, and both the `Wav.Chunks` and `Wav.Data` elements set to a blank data chunk.
func New(sampleRate uint32, bitDepth, numChannels, format uint16) (*Wav, error) {
	h, err := NewHeader(sampleRate, bitDepth, numChannels, format)
	if err != nil {
		return nil, err
	}

	return FromHeader(h)
}

// FromHeader creates a new Wav, configured with the input header.Header.
//
// This call returns a pointer to a Wav, and an error which is raised if the input
// header.Header is invalid.
//
// The returned Wav object will have its header set in every field except for
// `ChunkSize`, and both the `Wav.Chunks` and `Wav.Data` elements set to a blank data chunk.
func FromHeader(head *Header) (*Wav, error) {
	if head == nil {
		return nil, ErrEmptyHeader
	}

	if err := ValidateHeader(head); err != nil {
		return nil, err
	}

	blankData := NewChunk(nil, head.BitsPerSample, head.AudioFormat)

	return &Wav{
		Header: head,
		Chunks: []Chunk{blankData},
		Data:   blankData,
	}, nil
}

// From creates a new Wav, as read from the input io.Reader.
func From(r io.Reader) (w *Wav, err error) {
	w = new(Wav)

	if _, err = w.ReadFrom(r); err != nil {
		return nil, err
	}

	return w, nil
}

// Generate wraps a call to w.Data.Generate, by passing the same sample rate
// value as configured in w.header.SampleRate.
func (w *Wav) Generate(waveType osc.Type, freq int, dur time.Duration) {
	if w.Header == nil {
		return
	}

	w.Data.Generate(waveType, freq, int(w.Header.SampleRate), dur)
}
