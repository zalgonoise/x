// Package recorder persists audio as a series of encoded segments (CLAUDE.md §7 Q7): a Recorder is an sdk.Exporter that
// cuts the stream into segments of a fixed duration, encodes each one (WAV by default) and stores it as one object in
// a Store, such as a local directory or an S3-compatible bucket.
//
// Segments are encoded in memory, so every stored object is a complete file with exact sizes, and memory use is
// bounded by the segment duration (one minute of 48 kHz stereo 16-bit is about 11 MiB). Storing runs on the Export
// goroutine: wrap a Recorder in an sdk.QueuedExporter with the Block policy to record a live stream without loss.
package recorder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/internal/iox"
	"github.com/zalgonoise/x/audio/sdk"
)

// DefaultSegmentDuration is how much audio each segment holds unless WithSegmentDuration says otherwise.
const DefaultSegmentDuration = time.Minute

// Store keeps encoded segments. Put stores size bytes from r under name, replacing any object of that name. r is only
// valid until Put returns (the Recorder reuses its buffer): a Store that uploads asynchronously must copy it first.
type Store interface {
	Put(ctx context.Context, name string, r io.Reader, size int64) error
}

// Encoder encodes one segment's samples. It is a SampleWriter that Close completes.
type Encoder interface {
	audio.SampleWriter
	io.Closer
}

// EncoderFunc creates an Encoder writing a segment of format f to w. The default is NewWAV.
type EncoderFunc func(w io.Writer, f audio.Format) (Encoder, error)

// NewWAV is the default EncoderFunc: a WAV file in the source's own format.
func NewWAV(w io.Writer, f audio.Format) (Encoder, error) {
	enc, err := wav.NewEncoder(w, f)
	if err != nil {
		return nil, err
	}

	return enc, nil
}

// Segment describes a segment being named.
type Segment struct {
	Start      time.Time    // when the Recorder opened the segment (see WithClock)
	FirstFrame int64        // index of its first frame in the stream
	Format     audio.Format // its format
}

// NameFunc names a segment. The default is "<start, UTC, ms precision>_<first frame, 12 digits>.wav", for example
// "20261027T031500.250Z_000000441000.wav": names sort by time, and never collide when a stream restarts its frame
// numbering (a reconnect runs the pipeline again from frame 0).
type NameFunc func(s Segment) string

const startLayout = "20060102T150405.000Z"

func defaultName(s Segment) string {
	return fmt.Sprintf("%s_%012d.wav", s.Start.UTC().Format(startLayout), s.FirstFrame)
}

// Option configures a Recorder.
type Option func(*Recorder) error

// WithSegmentDuration sets how much audio each segment holds (at least one frame). The default is
// DefaultSegmentDuration.
func WithSegmentDuration(d time.Duration) Option {
	return func(r *Recorder) error {
		if d <= 0 {
			return fmt.Errorf("%w: segment duration %s", sdk.ErrInvalidOption, d)
		}

		r.segment = d

		return nil
	}
}

// WithEncoder sets how segments are encoded (for example FLAC, once it exists). The default is NewWAV.
func WithEncoder(fn EncoderFunc) Option {
	return func(r *Recorder) error {
		if fn == nil {
			return fmt.Errorf("%w: nil encoder", sdk.ErrInvalidOption)
		}

		r.encode = fn

		return nil
	}
}

// WithClock sets the clock that timestamps segments (Segment.Start). The default is time.Now.
func WithClock(now func() time.Time) Option {
	return func(r *Recorder) error {
		if now == nil {
			return fmt.Errorf("%w: nil clock", sdk.ErrInvalidOption)
		}

		r.now = now

		return nil
	}
}

// WithNames sets how segments are named. Names must suit the Store (Dir accepts local paths only), and should be
// unique: a Store replaces an object stored under the same name.
func WithNames(fn NameFunc) Option {
	return func(r *Recorder) error {
		if fn == nil {
			return fmt.Errorf("%w: nil name func", sdk.ErrInvalidOption)
		}

		r.name = fn

		return nil
	}
}

// Recorder is an sdk.Exporter storing the stream as segments of contiguous frames. A segment is stored when it is full,
// when the format changes, when a Buffer does not continue it (a gap, an overlap, or a stream that restarted from frame
// 0 after a reconnect), and on ForceFlush and Shutdown (which store the partial segment).
type Recorder struct {
	store   Store
	segment time.Duration
	encode  EncoderFunc
	name    NameFunc
	now     func() time.Time

	mu     sync.Mutex
	buf    iox.SeekBuffer
	enc    Encoder
	format audio.Format
	first  int64     // first frame of the open segment
	start  time.Time // when the open segment was opened
	frames int64     // frames in the open segment
	limit  int64     // frames per segment for the open segment's format
	closed bool
}

// New returns a Recorder storing segments in store.
func New(store Store, opts ...Option) (*Recorder, error) {
	if store == nil {
		return nil, fmt.Errorf("%w: nil store", sdk.ErrInvalidOption)
	}

	r := &Recorder{store: store, segment: DefaultSegmentDuration, encode: NewWAV, name: defaultName, now: time.Now}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		if err := opt(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

// Export implements sdk.Exporter. It splits b across segment boundaries as needed and stores every segment it
// completes; a failure to encode or store one is returned, and that segment is discarded.
func (r *Recorder) Export(ctx context.Context, b audio.Buffer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return sdk.ErrShutdown
	}

	if b.Frames() == 0 {
		return nil
	}

	var errs []error

	if r.enc != nil && (b.Format != r.format || b.Frame != r.first+r.frames) {
		errs = append(errs, r.finish(ctx)) // a new format, or audio that does not follow, starts a new segment
	}

	ch := b.Format.Channels
	frame := b.Frame

	for samples := b.Samples[:b.Frames()*ch]; len(samples) > 0; {
		if r.enc == nil {
			if err := r.open(b.Format, frame); err != nil {
				return errors.Join(append(errs, err)...)
			}
		}

		take := min(int64(len(samples)/ch), r.limit-r.frames)

		if _, err := r.enc.WriteSamples(samples[:take*int64(ch)]); err != nil {
			r.discard()

			return errors.Join(append(errs, fmt.Errorf("recorder: encoding: %w", err))...)
		}

		samples = samples[take*int64(ch):]
		frame += take
		r.frames += take

		if r.frames == r.limit {
			errs = append(errs, r.finish(ctx))
		}
	}

	return errors.Join(errs...)
}

// open starts a segment at frame.
func (r *Recorder) open(f audio.Format, frame int64) error {
	r.buf.Reset()

	enc, err := r.encode(&r.buf, f)
	if err != nil {
		return fmt.Errorf("recorder: creating an encoder: %w", err)
	}

	r.enc, r.format, r.first, r.frames, r.start = enc, f, frame, 0, r.now()
	r.limit = max(f.Frames(r.segment), 1)

	return nil
}

// finish completes the open segment, if any, and stores it.
func (r *Recorder) finish(ctx context.Context) error {
	if r.enc == nil {
		return nil
	}

	enc, name := r.enc, r.name(Segment{Start: r.start, FirstFrame: r.first, Format: r.format})
	r.enc = nil

	if err := enc.Close(); err != nil {
		return fmt.Errorf("recorder: finishing %s: %w", name, err)
	}

	data := r.buf.Bytes()
	if err := r.store.Put(ctx, name, bytes.NewReader(data), int64(len(data))); err != nil {
		return fmt.Errorf("recorder: storing %s: %w", name, err)
	}

	return nil
}

// discard drops the open segment after an encoding error.
func (r *Recorder) discard() {
	r.enc = nil
	r.buf.Reset()
}

// ForceFlush implements sdk.Exporter: it stores the partial segment; the next Export starts a new one.
func (r *Recorder) ForceFlush(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return sdk.ErrShutdown
	}

	return r.finish(ctx)
}

// Shutdown implements sdk.Exporter: it stores the partial segment. It is idempotent.
func (r *Recorder) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true

	return r.finish(ctx)
}
