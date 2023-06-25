package wav

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	dataheader "github.com/zalgonoise/x/audio/wav/data/header"
	"github.com/zalgonoise/x/audio/wav/header"
)

type multiProc struct {
	fns []func(float64) error
}

// MultiProc merges multiple processor functions for floating point audio data, with
// or without a fail-fast strategy
func MultiProc(failFast bool, fns ...func([]float64) error) func([]float64) error {
	if failFast {
		return func(data []float64) error {
			for i := range fns {
				if err := fns[i](data); err != nil {
					return err
				}
			}

			return nil
		}
	}

	return func(data []float64) error {
		var errs = make([]error, 0, len(fns))
		for i := range fns {
			if err := fns[i](data); err != nil {
				errs = append(errs, err)
			}
		}

		switch len(errs) {
		case 0:
			return nil
		case 1:
			return errs[0]
		default:
			return errors.Join(errs...)
		}
	}
}

// ByteRate calculates the byte rate of a certain signal from its header metadata
func ByteRate(sampleRate uint32, bitDepth, numChannels uint16) uint32 {
	return sampleRate * uint32(bitDepth) * uint32(numChannels) / 8
}

// TimeToBufferSize calculates the number of samples that are in a certain `dur` time.Duration,
// in the context of a byte-rate of `byteRate`
func TimeToBufferSize(byteRate uint32, dur time.Duration) (size int64) {
	rate := (int64)(time.Second) / (int64)(byteRate)

	return (int64)(dur) / rate
}

// RatioToBufferSize calculates the number of samples that are in a buffer,
// when a byte-rate of `byteRate` (that is equivalent to one second of audio) is multiplied by
// float64 `ratio`
func RatioToBufferSize(byteRate uint32, ratio float64) (size int64) {
	rate := (int64)(time.Second) / (int64)(byteRate)

	if ratio <= 0.0 {
		return rate
	}

	return int64(float64(rate) * ratio)
}

// Stream wraps a Wav type with custom functionality, allowing a ring-buffer approach
// to the data chunk (single-allocation float64 buffers), and optionally a processor function
type Stream struct {
	*Wav

	size int
	proc func([]float64) error
}

// NewStream creates a Stream with a certain size `size` and processor function `proc`
func NewStream(size int, proc func([]float64) error) *Stream {
	return &Stream{
		Wav:  new(Wav),
		size: size,
		proc: proc,
	}
}

// Stream reads the audio data in the io.Reader `r`, with the input context.
//
// Any errors raised either during reading the data or processing it are piped to the input
// errors channel `errCh`
func (w *Stream) Stream(ctx context.Context, r io.Reader, errCh chan<- error) {
	origFn := w.proc

	w.proc = func(data []float64) (err error) {
		if err = origFn(data); err != nil {
			errCh <- err
		}
		return err
	}

	if err := w.stream(ctx, r); err != nil {
		errCh <- err
		return
	}
}

func (w *Stream) stream(ctx context.Context, r io.Reader) (err error) {
	go w.ReadFrom(r)

	streamCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

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

// Write implements the io.Writer interface
//
// Write will gradually build a Wav object from the data passed through the
// slice of bytes `buf` input parameter. This method follows the lifetime of a
// Wav file from start to finish, even if it is raw and without a header.
//
// The method returns the number of bytes read by the buffer, and an error if the
// data is invalid (or too short)
func (w *Stream) Write(buf []byte) (n int, err error) {
	if w.readOnly.Load() {
		w.buf.Reset()
		w.readOnly.Store(false)
	}

	if w.buf == nil {
		w.buf = bytes.NewBuffer(buf)

		return w.decode()
	}

	if n, err = w.buf.Write(buf); err != nil {
		return n, err
	}

	return w.decode()
}

// ReadFrom implements the io.ReaderFrom interface
//
// # It allows for a Wav file (or stream) to be read and decoded into a data structure
//
// This implementation differs from a stream implementation of the Wav data structure, which
// would scope the stored PCM data in a ring buffer, both to save on storage / memory, and
// to only keep the last X bits of an audio stream (usually for analysis).
func (w *Stream) ReadFrom(r io.Reader) (n int64, err error) {
	var num int64

	if w.Header == nil {
		w.Header = new(header.Header)
	}

	if num, err = w.Header.ReadFrom(r); err != nil {
		return n, err
	}

	n += num

	for w.Data == nil {
		h := new(dataheader.Header)

		if num, err = h.ReadFrom(r); err != nil {
			return n, err
		}

		n += num

		chunk := NewRingChunk(h, w.Header.BitsPerSample, w.Header.AudioFormat, w.size, w.proc)
		w.Chunks = append(w.Chunks, chunk)

		if chunk.BitDepth() > 0 {
			w.Data = chunk
		}

		if num, err = chunk.ReadFrom(r); err != nil {
			return n, err
		}

		n += num
	}

	return n, nil
}

func (w *Stream) decode() (n int, err error) {
	if w.Header == nil {
		n, err = w.decodeHeader()
		if err != nil {
			return n, err
		}

		// header is required beyond this point, as w.head.BitsPerSample is necessary
		if w.Header == nil {
			return n, ErrMissingHeader
		}
	}

	for w.buf.Len() > 0 {
		if w.Data != nil {
			return w.decodeIntoData(n)
		}

		n, err = w.decodeNewSubChunk(n)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (w *Stream) decodeNewSubChunk(n int) (int, error) {
	// try to read subchunk headers
	if w.buf.Len() > dataheader.Size {
		var (
			err            error
			subchunk       *dataheader.Header
			subchunkBuffer = make([]byte, dataheader.Size)
		)

		if _, err = w.buf.Read(subchunkBuffer); err != nil {
			return n, err
		}

		if subchunk, err = dataheader.From(subchunkBuffer); err == nil {
			n += dataheader.Size
			chunk := NewRingChunk(subchunk, w.Header.BitsPerSample, w.Header.AudioFormat, w.size, w.proc)
			if string(subchunk.Subchunk2ID[:]) == dataSubchunkID {
				w.Data = chunk
			}

			end := int(subchunk.Subchunk2Size)
			ln := w.buf.Len()
			// grab remaining bytes if the byte slice isn't long enough
			// for a subchunk read
			if end > 0 && end > ln {
				end = ln - (ln % int(w.Header.BlockAlign))
			}

			chunkBuffer := make([]byte, end)
			if _, err = w.buf.Read(chunkBuffer); err != nil {
				return n, err
			}

			chunk.Parse(chunkBuffer)
			w.Chunks = append(w.Chunks, chunk)
			n += end
		}
	}
	return n, nil
}

func (w *Stream) decodeIntoData(n int) (int, error) {
	var (
		err error
		end = int(w.Data.Header().Subchunk2Size)
		ln  = w.buf.Len()
	)

	if end > 0 && end > ln {
		end = ln - (ln % int(w.Header.BlockAlign))
	}

	chunkBuffer := make([]byte, ln)
	if _, err = w.buf.Read(chunkBuffer); err != nil {
		return n, err
	}

	w.Data.Parse(chunkBuffer)
	return n + ln, nil
}
