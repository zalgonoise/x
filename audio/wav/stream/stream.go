package stream

import (
	"context"
	"io"

	"github.com/zalgonoise/x/audio/wav"

	"github.com/zalgonoise/gbuf"

	dataheader "github.com/zalgonoise/x/audio/wav/data/header"
	"github.com/zalgonoise/x/audio/wav/header"
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
	_, err := w.Data.Write(buf)
	if err != nil {
		return err
	}

	if len(w.Filters) == 0 {
		return nil
	}

	//defer w.Data.Reset()

	for _, fn := range w.Filters {
		if err = fn(w, buf); err != nil {
			return err
		}
	}

	return nil
}

func (w *Wav) streamingFunc(cancel context.CancelCauseFunc) {
	if n, err := w.ring.ReadFrom(w.Reader); err != nil {
		_ = n
		cancel(err)
		return
	}
	// end of stream, return EOF
	cancel(io.EOF)
}

func (w *Wav) stream(ctx context.Context) (err error) {
	streamCtx, cancel := context.WithCancelCause(ctx)
	w.done = cancel

	if w.Header == nil {
		w.Header = new(header.Header)
	}

	if _, err = w.Header.ReadFrom(w.Reader); err != nil {
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

	for w.Data == nil {
		h := new(dataheader.Header)

		if _, err = h.ReadFrom(w.Reader); err != nil {
			return err
		}

		chunk := wav.NewChunk(h, w.Header.BitsPerSample, w.Header.AudioFormat)
		w.Chunks = append(w.Chunks, chunk)

		if chunk.BitDepth() > 0 {
			w.Data = chunk
		}
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
			err = context.Cause(streamCtx)
			if err != nil {
				return err
			}
			return streamCtx.Err()
		default:
		}
	}
}
