package stream

import (
	"io"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/encoding/wav/header"

	"github.com/zalgonoise/gbuf"
)

// Wav is just like a wav.Wav, but it's designed to support WAV audio streams
// from an io.Writer.
//
// Besides sharing similar elements with a Wav object, it also stores a slice of
// StreamFilter that are applied on each pass of data through the gbuf.RingBuffer.
//
// Its stored reader is also a public element of WavBuffer so that it can be reused
// within a StreamFilter function.
type Wav struct {
	Header    *header.Header
	Chunks    []wav.Chunk
	Data      wav.Chunk
	Filters   []StreamFilter
	Reader    io.Reader
	ring      *gbuf.RingFilter[byte]
	ratio     float64
	blockSize int
	done      func(error)
}

// New uses the input io.Reader `r` to create a stream.Wav
func New(r io.Reader) *Wav {
	return &Wav{
		Reader: r,
		ratio:  1.0,
	}
}
