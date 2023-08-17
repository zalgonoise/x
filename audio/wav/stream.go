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

// ProcessFunc describes a function that processes a portion of the audio buffer
// as it is read and decoded from the incoming byte stream
type ProcessFunc func(header *header.Header, data []float64) error

// MultiProc merges multiple processor functions for floating point audio data, with
// or without a fail-fast strategy
func MultiProc(failFast bool, fns ...ProcessFunc) ProcessFunc {
	switch len(fns) {
	case 0:
		return nil
	case 1:
		return fns[0]
	}

	if failFast {
		return func(h *header.Header, data []float64) error {
			for i := range fns {
				if err := fns[i](h, data); err != nil {
					return err
				}
			}

			return nil
		}
	}

	return func(h *header.Header, data []float64) error {
		var errs = make([]error, 0, len(fns))
		for i := range fns {
			if err := fns[i](h, data); err != nil {
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
	return sampleRate * uint32(numChannels) * (uint32(bitDepth) / 8)
}

// TimeToBufferSize calculates the number of samples that are in a certain `dur` time.Duration,
// in the context of a byte-rate of `byteRate`
func TimeToBufferSize(byteRate uint32, dur time.Duration) (size int) {
	rate := (int)(time.Second) / (int)(byteRate)

	return (int)(dur) / rate
}

// RatioToBufferSize calculates the number of samples that are in a buffer,
// when a byte-rate of `byteRate` (that is equivalent to one second of audio) is multiplied by
// float64 `ratio`
func RatioToBufferSize(byteRate uint32, ratio float64) (size int) {
	rate := (int)(time.Second) / (int)(byteRate)

	if ratio <= 0.0 {
		return rate
	}

	return int(float64(rate) * ratio)
}

// Stream wraps a Wav type with custom functionality, allowing a ring-buffer approach
// to the data chunk (single-allocation float64 buffers), and optionally a processor function
type Stream struct {
	*Wav

	Size int
	proc ProcessFunc

	cfg *StreamConfig
}

// StreamConfig holds different configuration settings for a Stream
type StreamConfig struct {
	// Size describes different settings for the Stream's buffer size
	Size SizeConfig
}

// SizeConfig contains different configurations to define the Stream's buffer size
type SizeConfig struct {
	// Size is a concrete value for the Stream's buffer size (in bytes)
	Size int
	// Dur is a time.Duration value for the desired Stream buffer, that is later translated to a concrete value
	Dur time.Duration
	// Ratio is a float64 value describing a ratio against 1 second (e.g. 0.5 is half-a-second, 2.0 is two seconds)
	Ratio float64
}

// NewStream creates a Stream with a certain StreamConfig `cfg` and processor function `proc`
//
// The size is in bytes and can be calculated through one of the available *ToBufferSize functions
func NewStream(cfg *StreamConfig, proc ProcessFunc) *Stream {
	if cfg == nil {
		cfg = new(StreamConfig)
	}

	return &Stream{
		Wav:  new(Wav),
		cfg:  cfg,
		proc: proc,
	}
}

// Stream reads the audio data in the io.Reader `r`, with the input context.
//
// Any errors raised either during reading the data or processing it are piped to the input
// errors channel `errCh`
func (w *Stream) Stream(ctx context.Context, r io.Reader, errCh chan<- error) {
	origFn := w.proc

	w.proc = func(h *header.Header, data []float64) (err error) {
		if err = origFn(h, data); err != nil {
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
	streamCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	go func() {
		if _, readErr := w.ReadFrom(r); err != nil {
			cancel(readErr)
		}
	}()

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

// Head returns the Stream's Wav.Header, or it will set it from consuming the first header.Size bytes
// from the input io.Reader.
func (w *Stream) Head(r io.Reader) (*header.Header, error) {
	if w.Header == nil {
		h := new(header.Header)

		if _, err := h.ReadFrom(r); err != nil {
			return nil, err
		}

		w.Header = h
	}

	return w.Header, nil
}

// ReadFrom implements the io.ReaderFrom interface.
//
// It allows for a Wav file (or stream) to be read and decoded into a data structure.
//
// This implementation differs from a stream implementation of the Wav data structure, which
// would scope the stored PCM data in a ring buffer, both to save on storage / memory, and
// to only keep the last X bits of an audio stream (usually for analysis).
func (w *Stream) ReadFrom(r io.Reader) (n int64, err error) {
	var num int64

	if w.Header == nil {
		w.Header = new(header.Header)

		if num, err = w.Header.ReadFrom(r); err != nil {
			return n, err
		}

		n += num
	}

	// correct Stream.Size if it is off with the bit-depth in the signal
	w.checkSize()

	for w.Data == nil {
		h := new(dataheader.Header)

		if num, err = h.ReadFrom(r); err != nil {
			return n, err
		}

		n += num

		chunk := NewRingChunk(h, w.Header.BitsPerSample, w.Header.AudioFormat, w.Size, func(data []float64) error {
			return w.proc(w.Header, data)
		})

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

func (w *Stream) checkSize() {
	switch {
	case w.Header == nil:
	case w.cfg.Size.Ratio > 0.0:
		w.Size = RatioToBufferSize(ByteRate(
			w.Header.SampleRate, w.Header.BitsPerSample, w.Header.NumChannels,
		), w.cfg.Size.Ratio)

	case w.cfg.Size.Size > 0:
		w.Size = w.cfg.Size.Size

	case w.cfg.Size.Dur > 0:
		w.Size = TimeToBufferSize(ByteRate(
			w.Header.SampleRate, w.Header.BitsPerSample, w.Header.NumChannels,
		), w.cfg.Size.Dur)

	default:
		w.Size = int(ByteRate(
			w.Header.SampleRate, w.Header.BitsPerSample, w.Header.NumChannels,
		))
	}

	if w.Size < int(w.Header.BitsPerSample) {
		w.Size = int(w.Header.BitsPerSample)
	}

	if offset := w.Size % int(w.Header.BitsPerSample); offset > 0 {
		w.Size += int(w.Header.BitsPerSample) - offset
	}
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

	// correct Stream.Size if it is off with the bit-depth in the signal
	w.checkSize()

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

			chunk := NewRingChunk(subchunk, w.Header.BitsPerSample, w.Header.AudioFormat, w.Size, func(data []float64) error {
				return w.proc(w.Header, data)
			})

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

// Read implements the io.Reader interface
//
// Read will write to the input slice of bytes `buf` the contents
// of the Wav `w`.
//
// It returns the number of bytes written to the buffer, and an error if the buffer
// is not big enough
func (w *Stream) Read(buf []byte) (n int, err error) {
	if !w.readOnly.Load() {
		w.encode()
		w.readOnly.Store(true)
	}

	return w.buf.Read(buf)
}

// Bytes casts the contents of the Wav `w` as a slice of bytes, with WAV file encoding
func (w *Stream) Bytes() []byte {
	if !w.readOnly.Load() {
		w.encode()
		w.readOnly.Store(true)
	}

	return w.buf.Bytes()
}

func (w *Stream) encode() {
	var (
		n    int
		size = header.Size
	)

	for i := range w.Chunks {
		if w.Chunks[i].Header().Subchunk2ID == dataheader.Junk {
			size += dataheader.Size + int(w.Chunks[i].Header().Subchunk2Size)
			continue
		}

		size += dataheader.Size + w.Size
	}

	if w.Header.ChunkSize == 0 {
		w.Header.ChunkSize = uint32(size)
	}

	buf := make([]byte, size)
	_, _ = w.Header.Read(buf[n : n+header.Size])
	n += header.Size

	for i := range w.Chunks {
		var (
			chunkHeader = w.Chunks[i].Header()
			chunkSize   = int(chunkHeader.Subchunk2Size)
		)

		if w.Chunks[i].Header().Subchunk2ID == dataheader.Data && w.Size < chunkSize {
			chunkSize = w.Size
		}

		_, _ = chunkHeader.Read(buf[n : n+dataheader.Size])
		n += dataheader.Size
		_, _ = w.Chunks[i].Read(buf[n : n+chunkSize])
		n += chunkSize
	}

	w.readOnly.Store(true)
	w.buf = bytes.NewBuffer(buf)

	return
}
