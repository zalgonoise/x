package wav

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/audio/encoding/wav/data"
)

// ByteRate calculates the byte rate of a certain signal from its header metadata.
func ByteRate(sampleRate uint32, bitDepth, numChannels uint16) uint32 {
	return sampleRate * uint32(numChannels) * (uint32(bitDepth) / byteSize)
}

// TimeToBufferSize calculates the number of samples that are in a certain `dur` time.Duration,
// in the context of a byte-rate of `byteRate`.
func TimeToBufferSize(byteRate uint32, dur time.Duration) (size int) {
	rate := int(time.Second) / int(byteRate)

	return int(dur) / rate
}

// RatioToBufferSize calculates the number of samples that are in a buffer,
// when a byte-rate of `byteRate` (that is equivalent to one second of audio) is multiplied by
// float64 `ratio`.
func RatioToBufferSize(byteRate uint32, ratio float64) (size int) {
	rate := int(time.Second) / int(byteRate)

	if ratio <= 0.0 {
		return rate
	}

	return int(float64(rate) * ratio)
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

// Stream reads the audio data in the io.Reader `r`, with the input context.
//
// Any errors raised either during reading the data or processing it are piped to the input
// errors channel `errCh`.
func (w *Stream) Stream(ctx context.Context, r io.Reader, errCh chan<- error) {
	w.proc = ErrorPipeContext(w.proc, errCh)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		if _, readErr := w.ReadFrom(r); readErr != nil {
			errCh <- readErr
			close(errCh)
			cancel()
		}
	}()

	<-ctx.Done()
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
	var num int64

	if w.Header == nil {
		w.Header = new(Header)

		if num, err = w.Header.ReadFrom(r); err != nil {
			return n, err
		}

		n += num
	}

	// correct Stream.Size if it is off with the bit-depth in the signal
	w.checkSize()

	ctx := WithID(context.Background())

	for w.Data == nil {
		h := new(data.Header)

		if num, err = h.ReadFrom(r); err != nil {
			return n, err
		}

		n += num

		chunk := NewRingChunk(h, w.Header.BitsPerSample, w.Header.AudioFormat, w.Size, func(data []float64) error {
			return w.proc(ctx, w.Header, data)
		})

		w.Chunks = append(w.Chunks, chunk)

		if chunk.BitDepth() > 0 {
			w.Data = chunk
		}

		if w.cfg.hook != nil {
			r = NewReaderContextHook(ctx, w.Header, r, w.cfg.hook)
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
			return n, ErrEmptyHeader
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
	if w.buf.Len() < data.Size {
		return 0, ErrShortDataBuffer
	}

	var (
		err            error
		subchunk       *data.Header
		subchunkBuffer = make([]byte, data.Size)
		ctx            = WithID(context.Background())
	)

	if _, err = w.buf.Read(subchunkBuffer); err != nil {
		return n, err
	}

	if subchunk, err = data.From(subchunkBuffer); err == nil {
		n += data.Size

		chunk := NewRingChunk(subchunk, w.Header.BitsPerSample, w.Header.AudioFormat, w.Size, func(data []float64) error {
			return w.proc(ctx, w.Header, data)
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

	return n, nil
}

func (w *Stream) decodeIntoData(n int) (int, error) {
	var (
		err error
		end = int(w.Data.Header().Subchunk2Size)
		ln  = w.buf.Len()
	)

	if end > 0 && end > ln {
		ln -= ln % int(w.Header.BlockAlign)
	}

	chunkBuffer := make([]byte, ln)
	if _, err = w.buf.Read(chunkBuffer); err != nil {
		return n, err
	}

	w.Data.Parse(chunkBuffer)

	return n + ln, nil
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

	return w.buf.Read(buf)
}

// Bytes casts the contents of the Wav `w` as a slice of bytes, with WAV file encoding.
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
		size = Size
	)

	for i := range w.Chunks {
		if w.Chunks[i].Header().Subchunk2ID == data.JunkID {
			size += data.Size + int(w.Chunks[i].Header().Subchunk2Size)
			continue
		}

		size += data.Size + w.Size
	}

	if w.Header.ChunkSize == 0 {
		w.Header.ChunkSize = uint32(size)
	}

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
		n += chunkSize
	}

	w.readOnly.Store(true)
	w.buf = bytes.NewBuffer(buf)
}
