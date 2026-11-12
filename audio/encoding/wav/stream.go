package wav

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/audio/encoding/wav/data"
)

// ByteRate calculates the byte rate of a certain signal from its header metadata.
func ByteRate(sampleRate uint32, bitDepth, numChannels uint16) uint32 {
	return sampleRate * uint32(numChannels) * (uint32(bitDepth) / byteSize)
}

// minStreamSize is the smallest Stream buffer, in bytes; it matches the smallest ring the data package allocates.
const minStreamSize = 64

// TimeToBufferSize calculates the number of bytes in `dur` of audio at a byte-rate of `byteRate` (bytes per second).
// A non-positive duration returns 0.
func TimeToBufferSize(byteRate uint32, dur time.Duration) (size int) {
	if dur <= 0 {
		return 0
	}

	// split into whole seconds and the remainder so that long durations cannot overflow int64 nanoseconds * rate
	secs, frac := int64(dur/time.Second), int64(dur%time.Second)

	return int(int64(byteRate)*secs + int64(byteRate)*frac/int64(time.Second))
}

// RatioToBufferSize calculates the number of bytes in `ratio` seconds of audio at a byte-rate of `byteRate` (bytes per
// second): 1.0 is one second. A non-positive ratio returns one second's worth of bytes.
func RatioToBufferSize(byteRate uint32, ratio float64) (size int) {
	if ratio <= 0.0 {
		return int(byteRate)
	}

	return int(float64(byteRate) * ratio)
}

// Stream wraps a Wav type with custom functionality, allowing a ring-buffer approach
// to the data chunk (single-allocation float64 buffers), and optionally a processor function.
type Stream struct {
	*Wav

	Size int
	proc ProcessContextFunc

	cfg Config
}

// NewStream creates a Stream with a certain StreamConfig `cfg` and processor function `proc`.
//
// The size is in bytes and can be calculated through one of the available *ToBufferSize functions.
func NewStream(proc ProcessFunc, opts ...cfg.Option[Config]) *Stream {
	config := cfg.New(opts...)

	return &Stream{
		Wav:  new(Wav),
		cfg:  config,
		proc: processFuncWithContext(proc),
	}
}

// NewStreamContext creates a Stream with a certain StreamConfig `cfg` and context-based
// processor function `proc`.
//
// The size is in bytes and can be calculated through one of the available *ToBufferSize functions.
func NewStreamContext(proc ProcessContextFunc, opts ...cfg.Option[Config]) *Stream {
	config := cfg.New(opts...)

	return &Stream{
		Wav:  new(Wav),
		cfg:  config,
		proc: proc,
	}
}

// Stream reads the audio data in the io.Reader `r`, with the input context, until the input is exhausted, a read or
// processing error occurs, or ctx is done. It blocks until then.
//
// The processor receives ctx (with a stream ID added, see GetID), so values and cancellation set by the caller reach
// it; once ctx is done, processing stops at the next frame. At most one error is sent to `errCh`, which is owned by
// the caller and never closed; the send is abandoned if ctx is done first.
//
// Reading happens in a background goroutine so that Stream can return as soon as ctx is done even while the reader is
// blocked; that goroutine exits when the pending Read returns (close the reader, e.g. the HTTP response body, to
// unblock it).
func (w *Stream) Stream(ctx context.Context, r io.Reader, errCh chan<- error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		guard = new(procGuard)
		done  = make(chan error, 1)
	)

	go func() {
		_, err := w.readFrom(ctx, r, guard)
		done <- err
	}()

	var err error

	select {
	case err = <-done:
	case <-ctx.Done():
		// wait for a processor call already in flight (so the caller can safely shut its exporters down), but not
		// for a Read blocked on the source
		guard.stop()

		return
	}

	if err == nil || errCh == nil {
		return
	}

	select {
	case errCh <- err:
	case <-ctx.Done():
	}
}

// Write implements the io.Writer interface.
//
// Write will gradually build a Wav object from the data passed through the
// slice of bytes `buf` input parameter. This method follows the lifetime of a
// Wav file from start to finish, even if it is raw and without a header.
//
// The method returns the number of bytes read by the buffer, and an error if the
// data is invalid (or too short).
func (w *Stream) Write(buf []byte) (n int, err error) {
	return w.write(buf, w.decode)
}

// Head returns the Stream's Wav.Header, or it will set it from consuming the first header.Size bytes
// from the input io.Reader.
func (w *Stream) Head(r io.Reader) (*Header, error) {
	if w.Header == nil {
		h := new(Header)

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
	return w.readFrom(context.Background(), r, nil)
}

// procGuard serialises processor calls with Stream's return: once stopped, no processor call is running or will run.
type procGuard struct {
	mu      sync.Mutex
	stopped bool
}

func (g *procGuard) run(fn func() error) error {
	if g == nil {
		return fn()
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.stopped {
		return context.Canceled
	}

	return fn()
}

// stop waits for any processor call in flight and prevents further ones.
func (g *procGuard) stop() {
	g.mu.Lock()
	g.stopped = true
	g.mu.Unlock()
}

// readFrom implements ReadFrom with a context: the processor receives ctx (with a stream ID), and processing stops with
// ctx.Err() once ctx is done. Processor calls go through guard, when set.
func (w *Stream) readFrom(parent context.Context, r io.Reader, guard *procGuard) (n int64, err error) {
	var num int64

	ctx := WithID(parent)

	if w.Header == nil {
		header := new(Header)

		// the hook wraps the reader once, before anything is read, so it sees every byte exactly once
		if w.cfg.hook != nil {
			r = NewReaderContextHook(ctx, header, r, w.cfg.hook)
		}

		num, err = header.ReadFrom(r)
		n += num

		if err != nil {
			// never keep a half-decoded header: a retry would trust it
			return n, err
		}

		w.Header = header
	} else if w.cfg.hook != nil {
		r = NewReaderContextHook(ctx, w.Header, r, w.cfg.hook)
	}

	// correct Stream.Size if it is off with the frame size in the signal
	w.checkSize()

	for w.Data == nil {
		h := new(data.Header)

		num, err = h.ReadFrom(r)
		n += num

		if err != nil {
			return n, err
		}

		chunk := NewRingChunk(h, w.Header.BitsPerSample, w.Header.AudioFormat, w.Size, func(data []float64) error {
			return guard.run(func() error {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return ctxErr
				}

				return w.proc(ctx, w.Header, data)
			})
		})
		if chunk == nil {
			return n, fmt.Errorf("%w: BitsPerSample %d", ErrInvalidBitDepth, w.Header.BitsPerSample)
		}

		alignFrames(chunk, w.Header.BlockAlign)
		w.Chunks = append(w.Chunks, chunk)

		if chunk.BitDepth() > 0 {
			w.Data = chunk
		}

		num, err = chunk.ReadFrom(r)
		n += num

		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (w *Stream) checkSize() {
	switch {
	case w.Header == nil:
	case w.cfg.ratio > 0.0:
		w.Size = RatioToBufferSize(ByteRate(
			w.Header.SampleRate, w.Header.BitsPerSample, w.Header.NumChannels,
		), w.cfg.ratio)

	case w.cfg.size > 0:
		w.Size = w.cfg.size

	case w.cfg.dur > 0:
		w.Size = TimeToBufferSize(ByteRate(
			w.Header.SampleRate, w.Header.BitsPerSample, w.Header.NumChannels,
		), w.cfg.dur)

	default:
		w.Size = int(ByteRate(
			w.Header.SampleRate, w.Header.BitsPerSample, w.Header.NumChannels,
		))
	}

	// the buffer holds whole frames: round up to a multiple of BlockAlign (a validated header never has 0)
	align := max(int(w.Header.BlockAlign), 1)

	w.Size = max(w.Size, minStreamSize, align)

	if offset := w.Size % align; offset > 0 {
		w.Size += align - offset
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
			return n, ErrEmptyHeader
		}
	}

	// correct Stream.Size if it is off with the frame size in the signal
	w.checkSize()

	ctx := WithID(context.Background())

	return w.decodeChunks(n, func(h *data.Header) Chunk {
		return NewRingChunk(h, w.Header.BitsPerSample, w.Header.AudioFormat, w.Size, func(data []float64) error {
			return w.proc(ctx, w.Header, data)
		})
	})
}

// Read implements the io.Reader interface.
//
// Read will write to the input slice of bytes `buf` the contents
// of the Wav `w`.
//
// It returns the number of bytes written to the buffer, and an error if the buffer
// is not big enough.
func (w *Stream) Read(buf []byte) (n int, err error) {
	if !w.readOnly.Load() {
		w.encode()
		w.readOnly.Store(true)
	}

	return w.out.Read(buf)
}

// Bytes casts the contents of the Wav `w` as a slice of bytes, with WAV file encoding.
func (w *Stream) Bytes() []byte {
	if !w.readOnly.Load() {
		w.encode()
		w.readOnly.Store(true)
	}

	return w.out.Bytes()
}

func (w *Stream) encode() {
	var (
		n    int
		size = Size
	)

	for i := range w.Chunks {
		if w.Chunks[i].Header().Subchunk2ID != data.DataID {
			chunkSize := int(w.Chunks[i].Header().Subchunk2Size)
			size += data.Size + chunkSize + chunkSize&1 // odd-sized chunks are followed by a pad byte

			continue
		}

		size += data.Size + w.Size + w.Size&1 // odd-sized chunks are followed by a pad byte
	}

	// the RIFF size counts everything after the "RIFF" ID and the size field itself
	w.Header.ChunkSize = uint32(size - riffHeaderSize)

	buf := make([]byte, size)

	//nolint:errcheck // reading from the header should not raise any errors, and can be safely ignored.
	_, _ = w.Header.Read(buf[n : n+Size])
	n += Size

	for i := range w.Chunks {
		var (
			chunkHeader = w.Chunks[i].Header()
			chunkSize   = int(chunkHeader.Subchunk2Size)
		)

		if w.Chunks[i].Header().Subchunk2ID == data.DataID && w.Size < chunkSize {
			chunkSize = w.Size
		}

		//nolint:errcheck // reading from the chunk header should not raise any errors, and can be safely ignored.
		_, _ = chunkHeader.Read(buf[n : n+data.Size])
		n += data.Size

		//nolint:errcheck // reading from the chunk should not raise any errors, and can be safely ignored.
		_, _ = w.Chunks[i].Read(buf[n : n+chunkSize])
		n += chunkSize + chunkSize&1 // the pad byte is already zero
	}

	w.readOnly.Store(true)
	w.out = bytes.NewBuffer(buf)
}
