package wav

import (
	"context"
	"fmt"
	"io"

	"github.com/zalgonoise/gbuf"

	"github.com/zalgonoise/x/audio/wav/data"
)

const (
	minBufferSize  = 16
	baseBufferSize = 4
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
	Header    *WavHeader
	Chunks    []data.Chunk
	Data      data.Chunk
	Filters   []StreamFilter
	Reader    io.Reader
	ring      *gbuf.RingFilter[byte]
	ratio     float64
	blockSize int
	done      func(error)
}

// NewStream uses the input io.Reader `r` to create a WavBuffer
func NewStream(r io.Reader) *WavBuffer {
	return &WavBuffer{
		Reader: r,
		ratio:  1.0,
	}
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

func (w *WavBuffer) BlockSize(size int) *WavBuffer {
	if size < 0 {
		size = 0
	}
	w.blockSize = size
	return w
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

// Close implements the io.Closer interface
//
// It allows a graceful shutdown by cancelling the context in the WavBuffer
func (w *WavBuffer) Close() error {
	if w.done != nil {
		w.done(nil)
	}
	return nil
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
	defer w.Data.Reset()
	for _, fn := range w.Filters {
		if err := fn(w, buf); err != nil {
			return err
		}
	}
	return nil
}

func (w *WavBuffer) streamingFunc(cancel context.CancelCauseFunc) {
	if _, err := w.ring.ReadFrom(w.Reader); err != nil {
		cancel(err)
		return
	}

	// end of stream, return EOF
	cancel(io.EOF)
}

func (w *WavBuffer) stream(ctx context.Context) error {
	var (
		headerBuf         = make([]byte, 36) // fixed-size header
		subChunkBuf       = make([]byte, 8)  // fixed-size sub-chunk
		streamCtx, cancel = context.WithCancelCause(ctx)
	)
	w.done = cancel

	// read and parse header -- returns an error if raised AND header is unset
	if _, err := w.Reader.Read(headerBuf); err != nil && w.Header == nil {
		return err
	}
	if err := w.parseHeader(headerBuf); err != nil && w.Header == nil {
		return err
	}

	// define a buffer size based on the configured size, or calculate one from
	// the ratio, if it's over the minimum buffer size
	var bufferSize = w.blockSize

	if bufferSize < minBufferSize {
		bufferSize = int(w.Header.ByteRate)
		if float64(bufferSize)*w.ratio >= minBufferSize {
			bufferSize = int(float64(bufferSize) * w.ratio)
		}
	}

	// read and parse sub chunk section
	if _, err := w.Reader.Read(subChunkBuf); err != nil {
		return err
	}

	if err := w.parseSubChunk(subChunkBuf); err != nil {
		return err
	}

	// create a new Ring Buffer to continuously read from the stream in chunks,
	// with zero memory allocations
	//
	// Ring Buffer allows applying a filter function that process each chunk
	// emitted whenever the ring's head reaches the tail. Using the processChunk
	// method for it
	w.ring = gbuf.NewRingFilter(bufferSize, w.processChunk)

	// kick off a goroutine with the streaming function, where the Ring Buffer
	// reads from the WavBuffer's Reader. The context.CancelCauseFunc passed will
	// serve as a signal for when the stream has timed out or when an error is
	// raised
	go w.streamingFunc(cancel)

	// wait for a termination signal. this is expected to be a context.Done()
	// signal
	//
	// this signal will have an error (always); be it for deadline exceeded,
	// a read error, or, if the stream was finite, an io.EOF error. The error
	// is returned to the caller to be handled accordingly
	for {
		select {
		case <-streamCtx.Done():
			err := context.Cause(streamCtx)
			if err != nil {
				return err
			}
			return streamCtx.Err()
		default:
		}
	}
}

func (w *WavBuffer) encode() (int, [][]byte) {
	var (
		size      = baseBufferSize
		numChunks = len(w.Chunks)
		byteData  = make([][]byte, numChunks+1)
	)

	// set the first item in byteData to be the WavBuffer Header
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
