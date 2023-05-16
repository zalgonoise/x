package stream

import (
	"context"
	"io"

	"github.com/zalgonoise/gbuf"
)

const minBufferSize = 16

// Stream will kick off a stream read using the input context.Context `ctx` (for deadlines)
// and the error channel `errCh` (to send err to)
//
// While the Stream method is a blocking function, it is designed so it can be launched as a
// goroutine
func (w *Wav) Stream(ctx context.Context, errCh chan<- error) {
	if err := w.stream(ctx); err != nil {
		errCh <- err
		return
	}
}

// Close implements the io.Closer interface
//
// It allows a graceful shutdown by cancelling the context in the WavBuffer
func (w *Wav) Close() error {
	if w.done != nil {
		w.done(nil)
	}
	return nil
}

func (w *Wav) processChunk(buf []byte) error {
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

func (w *Wav) streamingFunc(cancel context.CancelCauseFunc) {
	if _, err := w.ring.ReadFrom(w.Reader); err != nil {
		cancel(err)
		return
	}

	// end of stream, return EOF
	cancel(io.EOF)
}

func (w *Wav) stream(ctx context.Context) error {
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
