package wav

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/zalgonoise/gbuf"

	"github.com/zalgonoise/x/audio/wav/data"
)

const (
	minBufferSize = 16
)

// WavBuffer is just like a Wav, but it's designed to support WAV audio streams
// from an io.Writer.
//
// Besides sharing similar elements with a Wav object, it also stores a slice of
// StreamFilter that are applied on each pass of data through the gbuf.RingBuffer.
//
// Its stored reader is also a public element of WavBuffer so that it can be reused
// within a StreamFilter function.
type WavBuffer struct {
	Header  *WavHeader
	Chunks  []data.Chunk
	Data    data.Chunk
	Filters []StreamFilter
	Reader  io.Reader
	ring    *gbuf.RingFilter[byte]
	ratio   float64
}

// Stream will kick off a stream read using the input context.Context `ctx` (for deadlines)
// and the error channel `errCh` (to send errors to)
//
// While the Stream method is a blocking function, it is designed so it can be launched as a
// goroutine
func (w *WavBuffer) Stream(ctx context.Context, errCh chan<- error) {
	if err := w.stream(ctx); err != nil {
		errCh <- err
		return
	}
}

// NewStream uses the input io.Reader `r` to create a WavBuffer
func NewStream(r io.Reader) *WavBuffer {
	return &WavBuffer{
		Reader: r,
		ratio:  1.0,
	}
}

// WithFilter appends the input slice of StreamFilter `fns` to the WavBuffer Filters,
// returning the same WavBuffer to allow chaining
func (w *WavBuffer) WithFilter(fns ...StreamFilter) *WavBuffer {
	for _, fn := range fns {
		if fn != nil {
			w.Filters = append(w.Filters, fn)
		}
	}
	return w
}

// Ratio sets the ring buffer's size ratio, short-circuiting if the input
// float64 `ratio` is zero; returning the same WavBuffer to allow chaining
func (w *WavBuffer) Ratio(ratio float64) *WavBuffer {
	if ratio == 0 {
		return w
	}
	w.ratio = ratio
	return w
}

func (w *WavBuffer) parseHeader(buf []byte) error {
	header, err := HeaderFrom(buf)
	if err != nil {
		return err
	}
	w.Header = header
	return nil
}

func (w *WavBuffer) parseSubChunk(buf []byte) error {
	subchunk, err := data.HeaderFrom(buf)
	if err != nil {
		return err
	}
	chunk := NewChunk(w.Header.BitsPerSample, subchunk)
	w.Chunks = append(w.Chunks, chunk)
	w.Data = chunk
	return nil
}

func (w *WavBuffer) processChunk(buf []byte) error {
	w.Data.Parse(buf)
	if len(w.Filters) == 0 {
		return nil
	}
	v := w.Data.Value()
	defer w.Data.Reset()
	for _, fn := range w.Filters {
		if err := fn(w, v, buf); err != nil {
			return err
		}
	}
	return nil
}

func (w *WavBuffer) stream(ctx context.Context) error {
	hbuf := make([]byte, 36)
	if _, err := w.Reader.Read(hbuf); err != nil {
		return err
	}

	if err := w.parseHeader(hbuf); err != nil && w.Header == nil {
		return err
	}

	bufferSize := int(w.Header.ByteRate)
	if float64(bufferSize)*w.ratio >= minBufferSize {
		bufferSize = int(float64(bufferSize) * w.ratio)
	}
	w.ring = gbuf.NewRingFilter(bufferSize, w.processChunk)
	scbuf := make([]byte, 8)
	if _, err := w.Reader.Read(scbuf); err != nil {
		return err
	}

	if err := w.parseSubChunk(scbuf); err != nil {
		return err
	}

	var err error
	go func() {
		if _, err = w.ring.ReadFrom(w.Reader); err != nil {
			return
		}
	}()

	for {
		select {
		case <-ctx.Done():
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			return nil
		default:
		}
	}
}

// Bytes casts the WavBuffer data as a WAV-file-encoded slice of bytes
func (w *WavBuffer) Bytes() []byte {
	var n int
	size, byteData := w.encode()

	buf := make([]byte, size+32)
	for i := range byteData {
		n += copy(buf[n:], byteData[i])
	}
	return buf
}

func (w *WavBuffer) encode() (size int, byteData [][]byte) {
	size = 4
	byteData = make([][]byte, 3)

	for i, j := 0, 1; i < len(w.Chunks); i, j = i+1, j+2 {
		byteData[j] = w.Chunks[i].Header().Bytes()
		byteData[j+1] = w.Chunks[i].Generate()
		size += 8 + len(byteData[j+1])
	}

	if w.Header.ChunkSize == 0 {
		w.Header.ChunkSize = uint32(size)
	}
	byteData[0] = w.Header.Bytes()
	return size, byteData
}

// Read implements the io.Reader interface
//
// It allows pushing the stored data to the input slice of bytes `buf`, returning
// the number of bytes written and an error if raised (if the input buffer is too short)
func (w *WavBuffer) Read(buf []byte) (n int, err error) {
	size, byteData := w.encode()
	if len(buf) < size {
		return n, fmt.Errorf("%w: input buffer with length %d cannot fit %d bytes", ErrShortDataBuffer, len(buf), size)
	}

	for i := range byteData {
		n += copy(buf[n:], byteData[i])
	}
	return size, nil
}
