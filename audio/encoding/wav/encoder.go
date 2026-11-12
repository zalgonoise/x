package wav

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/pcm"
	"github.com/zalgonoise/x/audio/encoding/riff"
)

// Errors returned by the Encoder.
var (
	// ErrClosed is returned by WriteSamples after Close.
	ErrClosed = errors.New("wav: encoder is closed")
	// ErrPartialFrame is returned by WriteSamples when len(src) is not a multiple of the channel count.
	ErrPartialFrame = errors.New("wav: samples do not fill whole frames")
	// ErrReservedChunk is returned for a metadata chunk the Encoder cannot write: fmt, data, or one without data.
	ErrReservedChunk = errors.New("wav: chunk cannot be written as metadata")
)

//nolint:gochecknoglobals // immutable identifier
var idFact = riff.ID("fact")

// EncoderOption configures an Encoder.
type EncoderOption func(*Encoder)

// WithChunks writes metadata chunks (LIST, bext, cue, ...) between the fmt chunk and the data chunk, in order, such as
// the Chunks of a Decoder when transcoding. fact chunks are dropped: the Encoder writes its own, with the right count,
// when the format needs one. fmt and data chunks, and chunks listed without their data (see WithMetadataLimit), make
// NewEncoder fail with ErrReservedChunk; chunks with an invalid ID, with riff.ErrBadID.
func WithChunks(chunks ...riff.Chunk) EncoderOption {
	return func(e *Encoder) {
		for _, c := range chunks {
			if c.ID != idFact {
				e.chunks = append(e.chunks, c)
			}
		}
	}
}

// Encoder writes samples as a RIFF/WAVE stream. It implements audio.SampleWriter.
//
// The fmt chunk follows Microsoft's guidance: WAVE_FORMAT_EXTENSIBLE (40 bytes) for PCM deeper than 16 bits, more than
// two channels, a channel mask or fewer significant bits than the container; otherwise PCMWAVEFORMAT (16 bytes) for PCM
// and WAVEFORMATEX (18 bytes) for float, A-law and µ-law. Non-PCM formats get a fact chunk with the frame count.
//
// When the destination can seek, Close back-patches the RIFF and data sizes and the fact count. Otherwise they are
// written as 0xFFFFFFFF, as live streams do (Decoder reads such streams to the end of the input), and an odd-length data
// chunk is left unpadded. The byte rate field saturates at 0xFFFFFFFF for formats beyond it; it is informational, and
// Decoder ignores it.
//
// Samples are quantised by package pcm: rounded, clipped to the integer range, NaN as silence. An Encoder is not safe for
// concurrent use.
type Encoder struct {
	riff   *riff.Writer
	data   io.Writer
	fact   *riff.Reservation
	format audio.Format
	codec  pcm.Codec
	chunks []riff.Chunk

	buf      []byte
	frames   int64
	closed   bool
	closeErr error
	err      error // the first write error
}

// NewEncoder writes the RIFF header, the fmt chunk, any metadata chunks and the data chunk header to w, and returns an
// Encoder for the samples. It returns an error matching audio.ErrInvalidFormat for an invalid Format, and one matching
// audio.ErrUnsupportedFormat for a format WAV cannot carry.
func NewEncoder(w io.Writer, f audio.Format, opts ...EncoderOption) (*Encoder, error) {
	if w == nil {
		return nil, fmt.Errorf("wav: %w", riff.ErrNilWriter)
	}

	codec, err := pcm.WAV(f)
	if err != nil {
		return nil, fmt.Errorf("wav: %w", err)
	}

	if f.Channels > math.MaxUint16 || f.BlockAlign() > math.MaxUint16 || int64(f.SampleRate) > math.MaxUint32 {
		return nil, fmt.Errorf("%w: wav: %s does not fit a fmt chunk", audio.ErrUnsupportedFormat, f)
	}

	e := &Encoder{format: f, codec: codec}
	for _, opt := range opts {
		opt(e)
	}

	for _, c := range e.chunks {
		if c.ID == idFmt || c.ID == idData || (c.Data == nil && c.Size > 0) {
			return nil, fmt.Errorf("%w: %q", ErrReservedChunk, c.ID.String())
		}

		if !c.ID.Valid() {
			return nil, fmt.Errorf("wav: %w: %q", riff.ErrBadID, c.ID[:])
		}
	}

	if err := e.writeHeader(w); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Encoder) writeHeader(w io.Writer) error {
	rw, err := riff.NewWriter(w, riff.WAVE)
	if err != nil {
		return err
	}

	if err := rw.WriteChunk(idFmt, appendFmt(nil, e.format)); err != nil {
		return err
	}

	if e.format.Encoding != audio.EncodingPCM {
		// the frame count, back-patched by Close when possible
		if e.fact, err = rw.Reserve(idFact, binary.LittleEndian.AppendUint32(nil, riff.UnknownSize)); err != nil {
			return err
		}
	}

	for _, c := range e.chunks {
		if err := rw.WriteChunk(c.ID, c.Data); err != nil {
			return err
		}
	}

	data, err := rw.WriteStream(idData)
	if err != nil {
		return err
	}

	e.riff, e.data = rw, data

	return nil
}

// appendFmt appends the body of the fmt chunk describing f, which must have passed NewEncoder's checks.
//
//nolint:gosec // NewEncoder bounds every field
func appendFmt(dst []byte, f audio.Format) []byte {
	le := binary.LittleEndian
	tag := formatTag(f.Encoding)
	extensible := f.Channels > 2 || f.ChannelMask != 0 || (f.ValidBits != 0 && f.ValidBits != f.BitDepth) ||
		(f.Encoding == audio.EncodingPCM && f.BitDepth > 16) // Microsoft's threshold for extensible PCM

	head := tag
	if extensible {
		head = ExtensibleFormat
	}

	dst = le.AppendUint16(dst, head)
	dst = le.AppendUint16(dst, uint16(f.Channels))
	dst = le.AppendUint32(dst, uint32(f.SampleRate))
	dst = le.AppendUint32(dst, uint32(min(int64(f.SampleRate)*int64(f.BlockAlign()), math.MaxUint32)))
	dst = le.AppendUint16(dst, uint16(f.BlockAlign()))
	dst = le.AppendUint16(dst, uint16(f.BitDepth))

	switch {
	case extensible:
		dst = le.AppendUint16(dst, extensibleCbSize)
		dst = le.AppendUint16(dst, uint16(f.Significant()))
		dst = le.AppendUint32(dst, f.ChannelMask)
		dst = le.AppendUint16(dst, tag)
		dst = append(dst, ksSubtypeSuffix...)
	case f.Encoding != audio.EncodingPCM:
		dst = le.AppendUint16(dst, 0) // cbSize of WAVEFORMATEX
	}

	return dst
}

func formatTag(e audio.Encoding) uint16 {
	switch e {
	case audio.EncodingFloat:
		return FloatFormat
	case audio.EncodingALaw:
		return ALawFormat
	case audio.EncodingMuLaw:
		return MuLawFormat
	default:
		return PCMFormat
	}
}

// Format returns the format of the samples the Encoder expects.
func (e *Encoder) Format() audio.Format { return e.format }

// Frames returns the number of frames written so far.
func (e *Encoder) Frames() int64 { return e.frames }

// WriteSamples encodes interleaved, normalised samples; see audio.SampleWriter. len(src) must be a multiple of the
// channel count, or nothing is written and ErrPartialFrame is returned. A write error is returned by every later call.
func (e *Encoder) WriteSamples(src []float64) (int, error) {
	switch {
	case e.closed:
		return 0, ErrClosed
	case e.err != nil:
		return 0, e.err
	case len(src)%e.format.Channels != 0:
		return 0, fmt.Errorf("%w: %d samples for %d channels", ErrPartialFrame, len(src), e.format.Channels)
	}

	e.buf = e.codec.AppendEncode(e.buf[:0], src)

	n, err := e.data.Write(e.buf)
	frames := n / e.format.BlockAlign()
	e.frames += int64(frames)

	if err != nil {
		e.err = err

		return frames * e.format.Channels, err
	}

	return len(src), nil
}

// Close finishes the stream: when the destination can be patched (see riff.Writer.Close), it back-patches the sizes and
// the fact count and pads the data chunk to an even length. It does not close the destination. It returns any write
// error a WriteSamples call returned, so checking Close is enough to know the file is complete; calling it again
// returns the same result.
func (e *Encoder) Close() error {
	if e.closed {
		return e.closeErr
	}

	e.closed = true

	if e.fact != nil {
		count := uint32(min(e.frames, math.MaxUint32)) //nolint:gosec // clamped
		if err := e.fact.Set(binary.LittleEndian.AppendUint32(nil, count)); err != nil {
			e.closeErr = err

			return err
		}
	}

	// a write error from WriteSamples is latched by the riff.Writer too, which returns it from Close
	e.closeErr = e.riff.Close()

	return e.closeErr
}
