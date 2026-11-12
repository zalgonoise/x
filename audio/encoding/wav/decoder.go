package wav

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/pcm"
	"github.com/zalgonoise/x/audio/encoding/riff"
	"github.com/zalgonoise/x/audio/internal/iox"
)

// WAVE format tags (wFormatTag) beyond PCMFormat and FloatFormat.
const (
	ALawFormat       uint16 = 0x0006
	MuLawFormat      uint16 = 0x0007
	ExtensibleFormat uint16 = 0xFFFE
)

const (
	fmtMinSize        = 16 // WAVEFORMAT plus wBitsPerSample (PCMWAVEFORMAT)
	fmtExtensibleSize = 40 // WAVEFORMATEXTENSIBLE
	extensibleCbSize  = 22 // cbSize of WAVEFORMATEXTENSIBLE

	// DefaultMetadataLimit is how many bytes of non-audio chunks a Decoder keeps unless WithMetadataLimit says otherwise.
	DefaultMetadataLimit = 1 << 20
	// ChunkOverhead is what listing a chunk costs against the metadata limit, besides its data: a riff.Chunk in a
	// growing slice.
	ChunkOverhead = 64
)

//nolint:gochecknoglobals // immutable identifiers
var (
	idFmt  = riff.ID("fmt ")
	idData = riff.ID("data")

	// ksSubtypeSuffix is what follows the 16-bit format tag in the KSDATAFORMAT_SUBTYPE_* GUIDs
	// {0000xxxx-0000-0010-8000-00AA00389B71} of a WAVE_FORMAT_EXTENSIBLE fmt chunk.
	ksSubtypeSuffix = []byte{0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x80, 0x00, 0x00, 0xAA, 0x00, 0x38, 0x9B, 0x71}
)

// Errors returned by the Decoder. Malformed input also matches audio.ErrInvalidFormat, and well-formed input that no codec
// handles matches audio.ErrUnsupportedFormat.
var (
	// ErrNotWAV is returned for input that is not a RIFF/WAVE stream.
	ErrNotWAV = errors.New("wav: not a RIFF/WAVE stream")
	// ErrNoFormat is returned when no fmt chunk precedes the data chunk.
	ErrNoFormat = fmt.Errorf("%w: wav: no fmt chunk before the data chunk", audio.ErrInvalidFormat)
	// ErrNoData is returned when the input ends before a data chunk.
	ErrNoData = fmt.Errorf("%w: wav: no data chunk", audio.ErrInvalidFormat)
)

// DecoderOption configures a Decoder.
type DecoderOption func(*Decoder)

// WithMetadataLimit sets how many bytes of non-audio chunks (LIST, fact, bext, cue, ...) a Decoder keeps in memory in
// total, counting ChunkOverhead bytes for each listed chunk besides its data. A chunk whose data does not fit is listed
// without its data; once even the overhead does not fit, it is not listed. A limit of 0 or less keeps nothing. The
// default is DefaultMetadataLimit.
func WithMetadataLimit(n int) DecoderOption {
	return func(d *Decoder) { d.budget = max(n, 0) }
}

// Decoder reads a RIFF/WAVE stream and decodes its samples. It implements audio.SampleReader.
//
// It accepts any sample rate and channel count, 8/16/24/32-bit integer PCM (8-bit is unsigned), 32/64-bit IEEE float,
// 8-bit A-law and µ-law, in plain (16/18-byte fmt) or WAVE_FORMAT_EXTENSIBLE (40-byte fmt) files. Samples are normalised
// by 2^(n-1) (see package pcm); ValidBits is reported, and needs no special handling because WAV left-justifies samples
// in their container.
//
// Chunks are walked by size, honouring pad bytes (see riff.Reader). The data chunk is bounded by its declared size, or
// runs to the end of the input when that size is 0xFFFFFFFF, or 0 without a definite RIFF size declaring chunks after
// it, as live streams write it. Reads tolerate any split of the input: partial frames are carried over to the next
// read, and a partial frame at the end of the input is dropped. Errors from the input, including the
// io.ErrUnexpectedEOF of a dropped connection, are returned as such: io.EOF only ever means the end of the audio.
//
// A Decoder is not safe for concurrent use.
type Decoder struct {
	riff   *riff.Reader
	data   io.Reader // body of the data chunk
	format audio.Format
	codec  pcm.Codec
	block  int   // bytes per frame
	total  int64 // frames in the data chunk as declared, or -1 when unbounded

	buf   []byte // raw sample bytes: carried bytes first, then fresh reads
	carry int    // bytes of a partial frame at the start of buf

	chunks     []riff.Chunk
	budget     int  // metadata bytes still allowed
	trailer    bool // ReadTrailer has run
	trailerErr error
	err        error // sticky: io.EOF once the audio ended, or the error that stopped decoding
}

// NewDecoder reads the RIFF header and every chunk up to the start of the data chunk, and returns a Decoder positioned
// at the first sample. It returns ErrNotWAV for input that is not RIFF/WAVE, an error matching audio.ErrInvalidFormat
// for a malformed header, and one matching audio.ErrUnsupportedFormat for a format no codec handles.
func NewDecoder(r io.Reader, opts ...DecoderOption) (*Decoder, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: nil reader", ErrNotWAV)
	}

	d := &Decoder{budget: DefaultMetadataLimit, total: -1}
	for _, opt := range opts {
		opt(d)
	}

	rr, err := riff.NewReader(r, idData)

	switch {
	case errors.Is(err, riff.ErrNotRIFF):
		return nil, fmt.Errorf("%w: %w", ErrNotWAV, err)
	case err != nil:
		return nil, err // an I/O error
	}

	if rr.Form() != riff.WAVE {
		return nil, fmt.Errorf("%w: RIFF form %q", ErrNotWAV, rr.Form().String())
	}

	d.riff = rr

	var haveFmt bool

	for {
		h, body, err := rr.Next()
		if err != nil {
			return nil, headerError(err, haveFmt)
		}

		switch h.ID {
		case idFmt:
			if haveFmt {
				return nil, fmt.Errorf("%w: wav: second fmt chunk", audio.ErrInvalidFormat)
			}

			if err := d.readFmt(body); err != nil {
				return nil, err
			}

			haveFmt = true
		case idData:
			if !haveFmt {
				return nil, ErrNoFormat
			}

			d.data = body
			if !rr.Unbounded() {
				d.total = int64(h.Size) / int64(d.block)
			}

			return d, nil
		default:
			if err := d.keep(h, body); err != nil {
				return nil, err
			}
		}
	}
}

// headerError classifies an error from walking the chunks before the data chunk.
func headerError(err error, haveFmt bool) error {
	switch {
	case errors.Is(err, io.EOF) && haveFmt:
		return ErrNoData
	case errors.Is(err, io.EOF):
		return ErrNoFormat
	case errors.Is(err, riff.ErrTruncated), errors.Is(err, riff.ErrBadID):
		return fmt.Errorf("%w: wav: %w", audio.ErrInvalidFormat, err)
	default:
		return err // an I/O error
	}
}

// Format returns the format of the decoded samples.
func (d *Decoder) Format() audio.Format { return d.format }

// Frames returns the number of frames the data chunk declares. ok is false for live streams, whose data size is unknown.
// A truncated file holds fewer frames than declared.
func (d *Decoder) Frames() (frames int64, ok bool) { return d.total, d.total >= 0 }

// Chunks returns the chunks other than fmt and data that the Decoder kept (see WithMetadataLimit), in file order: those
// before the data chunk, followed by those after it once ReadTrailer has run. The slices are owned by the Decoder.
func (d *Decoder) Chunks() []riff.Chunk { return d.chunks }

// ReadSamples decodes up to len(dst) interleaved samples into dst; see audio.SampleReader. It reads from the input at
// most once per whole frame it still needs, returning as soon as at least one whole frame is available. It returns
// io.ErrShortBuffer when dst cannot hold a single frame.
func (d *Decoder) ReadSamples(dst []float64) (int, error) {
	if d.err != nil {
		return 0, d.err
	}

	frames := len(dst) / d.format.Channels
	if frames == 0 {
		if len(dst) == 0 {
			return 0, nil
		}

		return 0, io.ErrShortBuffer
	}

	need := frames * d.block
	if cap(d.buf) < need {
		buf := make([]byte, need)
		copy(buf, d.buf[:d.carry])
		d.buf = buf
	}

	buf := d.buf[:need]

	// carry < block, so this waits for at least one whole frame. iox keeps the end of the input (io.EOF) apart from a
	// dropped connection (io.ErrUnexpectedEOF from the input itself).
	n, err := iox.ReadAtLeast(d.data, buf[d.carry:], d.block-d.carry)
	total := d.carry + n
	whole := total / d.block * d.block
	got := d.codec.Decode(dst, buf[:whole])
	d.carry = copy(buf, buf[whole:total])

	switch {
	case err == nil:
		return got, nil
	case errors.Is(err, io.EOF):
		d.err, d.carry = io.EOF, 0 // a partial frame at the end of the input is dropped

		return got, io.EOF
	default:
		d.err = err

		return got, err
	}
}

// ReadTrailer discards whatever audio has not been read and reads the chunks that follow the data chunk (some writers
// put LIST or id3 metadata there), adding them to Chunks; stray fmt and data chunks there are skipped. It stops at the
// end of the RIFF body when the RIFF size is definite, leaving any following bytes (another file, an ID3v1 tag,
// padding) unread. Afterwards ReadSamples returns
// io.EOF, or the error ReadTrailer returned, which every later call returns too. A live stream's data chunk runs to
// the end of the input, so it has no trailer and ReadTrailer returns at once.
func (d *Decoder) ReadTrailer() error {
	switch {
	case d.trailer:
		return d.trailerErr
	case d.err != nil && !errors.Is(d.err, io.EOF):
		return d.err
	}

	d.trailer = true
	d.err, d.carry = io.EOF, 0

	if d.trailerErr = d.readTrailer(); d.trailerErr != nil {
		d.err = d.trailerErr
	}

	return d.trailerErr
}

func (d *Decoder) readTrailer() error {
	for {
		h, body, err := d.riff.Next()

		switch {
		case errors.Is(err, io.EOF):
			return nil
		case errors.Is(err, riff.ErrTruncated), errors.Is(err, riff.ErrBadID):
			return fmt.Errorf("%w: wav: after the data chunk: %w", audio.ErrInvalidFormat, err)
		case err != nil:
			return err
		}

		if h.ID == idFmt || h.ID == idData {
			continue // a stray fmt or data chunk after the audio is not metadata, and cannot be written back
		}

		if err := d.keep(h, body); err != nil {
			return err
		}
	}
}

// keep records a non-audio chunk within the metadata budget. Its body is skipped by the riff.Reader if not read here.
func (d *Decoder) keep(h riff.Header, body io.Reader) error {
	if d.budget < ChunkOverhead {
		return nil
	}

	d.budget -= ChunkOverhead
	c := riff.Chunk{Header: h}

	if int64(h.Size) <= int64(d.budget) {
		data := make([]byte, h.Size)

		n, err := iox.ReadAtLeast(body, data, len(data))
		if err != nil && !errors.Is(err, io.EOF) { // io.EOF: a chunk truncated by the end of the input
			return err
		}

		c.Data = data[:n]
		d.budget -= n
	}

	d.chunks = append(d.chunks, c)

	return nil
}

// readFmt reads a fmt chunk and selects the codec.
func (d *Decoder) readFmt(body io.Reader) error {
	var b [fmtExtensibleSize]byte

	n, err := iox.ReadAtLeast(body, b[:], len(b))
	if err != nil && !errors.Is(err, io.EOF) { // io.EOF: a fmt chunk shorter than 40 bytes, checked by ParseFormat
		return err
	}

	f, err := ParseFormat(b[:n])
	if err != nil {
		return err
	}

	codec, err := pcm.WAV(f)
	if err != nil {
		return fmt.Errorf("wav: %w", err)
	}

	d.format, d.codec, d.block = f, codec, f.BlockAlign()

	return nil
}

// ParseFormat decodes the body of a fmt chunk: WAVEFORMAT/PCMWAVEFORMAT (16 bytes), WAVEFORMATEX (18 bytes or more) or
// WAVEFORMATEXTENSIBLE (40 bytes or more). Bytes beyond the 40th are ignored. Structural problems (zero channels, a
// block alignment that does not match the channels and bit depth, a truncated chunk) return an error matching
// audio.ErrInvalidFormat; a format tag or extensible sub-format other than PCM, IEEE float, A-law or µ-law returns one
// matching audio.ErrUnsupportedFormat. It does not check that a codec exists for the bit depth (see pcm.WAV).
//
// The byte rate field is not checked, since writers often get it wrong and it carries no information of its own.
func ParseFormat(b []byte) (audio.Format, error) {
	le := binary.LittleEndian

	if len(b) < fmtMinSize {
		return audio.Format{}, fmt.Errorf("%w: wav: fmt chunk of %d bytes", audio.ErrInvalidFormat, len(b))
	}

	var (
		tag        = le.Uint16(b[0:2])
		channels   = int(le.Uint16(b[2:4]))
		rate       = int(le.Uint32(b[4:8]))
		blockAlign = int(le.Uint16(b[12:14]))
		bits       = int(le.Uint16(b[14:16]))
		validBits  int
		mask       uint32
	)

	if tag == ExtensibleFormat {
		if len(b) < fmtExtensibleSize || le.Uint16(b[16:18]) < extensibleCbSize {
			return audio.Format{}, fmt.Errorf("%w: wav: WAVE_FORMAT_EXTENSIBLE fmt chunk of %d bytes", audio.ErrInvalidFormat,
				len(b))
		}

		validBits = int(le.Uint16(b[18:20]))
		mask = le.Uint32(b[20:24])

		if !bytes.Equal(b[26:40], ksSubtypeSuffix) {
			return audio.Format{}, fmt.Errorf("%w: wav: WAVE_FORMAT_EXTENSIBLE sub-format %X", audio.ErrUnsupportedFormat,
				b[24:40])
		}

		tag = le.Uint16(b[24:26])
	}

	// compressed formats (ADPCM, MP3, GSM...) are well-formed with their own block and bit rules: classify the tag
	// before applying the PCM ones
	enc := encodingOf(tag)
	if enc == audio.EncodingUnknown {
		return audio.Format{}, fmt.Errorf("%w: wav: format tag %#04x", audio.ErrUnsupportedFormat, tag)
	}

	if channels == 0 || rate == 0 || bits == 0 || validBits > bits {
		return audio.Format{}, fmt.Errorf("%w: wav: %d channels, %d Hz, %d (%d valid) bits per sample",
			audio.ErrInvalidFormat, channels, rate, bits, validBits)
	}

	// a sample occupies whole bytes; PCMWAVEFORMAT stores e.g. 12-bit samples in 16-bit containers
	container := (bits + 7) / 8 * 8 //nolint:mnd // round up to whole bytes
	if blockAlign != channels*container/8 {
		return audio.Format{}, fmt.Errorf("%w: wav: block align %d for %d channels of %d bits", audio.ErrInvalidFormat,
			blockAlign, channels, bits)
	}

	if validBits == 0 {
		validBits = bits
	}

	if validBits == container {
		validBits = 0
	}

	f := audio.Format{
		SampleRate:  rate,
		Channels:    channels,
		BitDepth:    container,
		ValidBits:   validBits,
		Encoding:    enc,
		ChannelMask: mask,
	}

	if err := f.Validate(); err != nil {
		return audio.Format{}, fmt.Errorf("wav: %w", err)
	}

	return f, nil
}

func encodingOf(tag uint16) audio.Encoding {
	switch tag {
	case PCMFormat:
		return audio.EncodingPCM
	case FloatFormat:
		return audio.EncodingFloat
	case ALawFormat:
		return audio.EncodingALaw
	case MuLawFormat:
		return audio.EncodingMuLaw
	default:
		return audio.EncodingUnknown
	}
}

//nolint:gochecknoinits // registers the format with audio.Decode, as image codecs do with image.RegisterFormat
func init() {
	audio.RegisterFormat("wav", "RIFF????WAVE", func(r io.Reader) (audio.SampleReader, error) {
		d, err := NewDecoder(r)
		if err != nil {
			return nil, err
		}

		return d, nil
	})
}
