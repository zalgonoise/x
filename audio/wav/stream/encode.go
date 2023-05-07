package stream

import (
	"fmt"
	"time"

	"github.com/zalgonoise/x/audio/osc"
	"github.com/zalgonoise/x/audio/wav"
)

const baseBufferSize = 4

// Generate wraps a call to w.Data.Generate, by passing the same sample rate
// value as configured in w.header.SampleRate
func (w *Wav) Generate(waveType osc.Type, freq int, dur time.Duration) {
	w.Data.Generate(waveType, freq, int(w.Header.SampleRate), dur)
}

// Read implements the io.Reader interface
//
// It allows pushing the stored data to the input slice of bytes `buf`, returning
// the number of bytes written and an error if raised (if the input buffer is too short)
func (w *Wav) Read(buf []byte) (n int, err error) {
	size, byteData := w.encode()
	if len(buf) < size {
		return n, fmt.Errorf("%w: input buffer with length %d cannot fit %d bytes", wav.ErrShortDataBuffer, len(buf), size)
	}

	for i := range byteData {
		n += copy(buf[n:], byteData[i])
	}
	return size, nil
}

// Bytes casts the WavBuffer data as a WAV-file-encoded slice of bytes
func (w *Wav) Bytes() []byte {
	var n int
	size, byteData := w.encode()

	buf := make([]byte, size+32)
	for i := range byteData {
		n += copy(buf[n:], byteData[i])
	}
	return buf
}

func (w *Wav) encode() (int, [][]byte) {
	var (
		size      = baseBufferSize
		numChunks = len(w.Chunks)
		byteData  = make([][]byte, numChunks+1)
	)

	// set the first item in byteData to be the WavBuffer header
	byteData[0] = w.Header.Bytes()

	// for each chunk, align a slice for the header, and another for the data
	// index `i` is for the WavBuffer Chunks, while index `j` (starting on 1)
	// is for the byteData slice index
	for i, j := 0, 1; i < numChunks; i, j = i+1, j+2 {
		byteData[j] = w.Chunks[i].Header().Bytes()
		byteData[j+1] = w.Chunks[i].Bytes()
		size += 8 + len(byteData[j+1]) // increment size, header is a fixed len
	}

	// update size if needed
	if w.Header.ChunkSize < uint32(size) {
		w.Header.ChunkSize = uint32(size)
	}

	return size, byteData
}
