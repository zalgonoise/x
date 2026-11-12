// Package riff reads and writes RIFF container files (the format behind WAV, AVI and WebP): a "RIFF" header naming a
// form type (such as "WAVE"), followed by a sequence of chunks, each a four-character ID, a little-endian 32-bit size
// and a body padded to an even length.
//
// The Reader walks chunks from any io.Reader without trusting sizes for allocations, skipping unread bytes and pad
// bytes for the caller; it tolerates the 0 and 0xFFFFFFFF sizes live streams use for their last chunk. The Writer pads
// odd-sized chunks and fixes up sizes when its destination is an io.WriteSeeker.
//
// This package is the container layer for the encoding/wav rebuild (CLAUDE.md §6, Phase 2); it knows nothing about
// audio.
package riff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/zalgonoise/x/audio/internal/iox"
)

// FourCC is a four-character chunk or form identifier.
type FourCC [4]byte

// String returns the identifier as text.
func (f FourCC) String() string { return string(f[:]) }

// Valid reports whether every character is printable ASCII, as RIFF requires (IDs are space-padded).
func (f FourCC) Valid() bool {
	for _, c := range f {
		if c < ' ' || c > '~' {
			return false
		}
	}

	return true
}

// ID returns the FourCC for a string of up to four characters, space-padded. Longer strings are truncated.
func ID(s string) FourCC {
	f := FourCC{' ', ' ', ' ', ' '}
	copy(f[:], s)

	return f
}

// Well-known identifiers.
//
//nolint:gochecknoglobals // immutable identifiers, used in comparisons
var (
	RIFF = ID("RIFF")
	WAVE = ID("WAVE")
)

const (
	headerSize = 8
	idSize     = 4 // a FourCC, and also the width of a size field
	// UnknownSize is the chunk size live streams write when the length is not known in advance.
	UnknownSize = math.MaxUint32
)

// Errors returned by the Reader and the Writer.
var (
	ErrNotRIFF   = errors.New("riff: not a RIFF file")
	ErrBadID     = errors.New("riff: invalid chunk ID")
	ErrTruncated = errors.New("riff: truncated chunk header")
	ErrClosed    = errors.New("riff: writer is closed")
	ErrSize      = errors.New("riff: size mismatch")
	ErrNilWriter = errors.New("riff: nil writer")
	// ErrAppendOnly is returned by Close when sizes cannot be patched because the destination appends every write and
	// cannot be truncated back after the probe write, or because it changed size during Close (another writer).
	ErrAppendOnly = errors.New("riff: cannot patch sizes: destination appends writes or changed size")
)

// Header is a chunk header.
type Header struct {
	ID   FourCC
	Size uint32 // as declared; may be 0 or UnknownSize for a live stream's final chunk
}

// Chunk is a whole chunk: its header and its body. Data may be nil when the reader chose not to keep the body (for
// example to bound memory), in which case Size still tells how large it was.
type Chunk struct {
	Header

	Data []byte
}

// Unbounded reports whether the size field holds one of the markers live streams write when the length is not known
// (0 or 0xFFFFFFFF). Whether a Reader takes such a chunk to run to the end of the input also depends on the RIFF size:
// see Reader.Unbounded.
func (h Header) Unbounded() bool { return h.Size == 0 || h.Size == UnknownSize }

// Padded reports whether a pad byte follows the chunk body (its declared size is odd).
func (h Header) Padded() bool { return !h.Unbounded() && h.Size%2 == 1 }

// Reader walks the chunks of a RIFF file.
//
// A definite RIFF size (anything but 0 and 0xFFFFFFFF) bounds the walk: once a stream chunk has been returned, Next
// reports io.EOF at the end of the RIFF body without reading past it, so concatenated files, appended ID3v1 tags and
// trailing padding are left alone. Before the first stream chunk the RIFF size is not trusted, so a wrong size cannot
// hide the audio.
//
// The Reader never mistakes an error from the underlying reader (such as the io.ErrUnexpectedEOF of a dropped HTTP
// connection) for the end of the input, and a reader that keeps returning (0, nil) makes it fail with
// io.ErrNoProgress.
type Reader struct {
	r    io.Reader
	form FourCC
	size uint32
	pos  int64 // bytes consumed after the RIFF size field

	cur     *io.LimitedReader // body of the current chunk
	curSize int64             // declared size of the current chunk
	pad     bool              // the current chunk is followed by a pad byte
	done    bool              // an unbounded chunk was returned: nothing can follow it
	stream  bool              // a stream chunk was returned
	streams map[FourCC]bool   // chunk IDs whose 0 / 0xFFFFFFFF size means "until EOF"
	err     error             // sticky: the walk cannot resume after a failure
}

// NewReader reads the RIFF header and returns a Reader positioned at the first chunk. Chunks listed in unbounded (for
// WAV: "data") are treated as extending to the end of the input when their declared size is 0xFFFFFFFF, or 0 unless a
// definite RIFF size declares more chunks after them.
//
// Input that is too short or does not start with "RIFF" returns ErrNotRIFF; a read error is returned as is.
func NewReader(r io.Reader, unbounded ...FourCC) (*Reader, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: nil reader", ErrNotRIFF)
	}

	r = iox.Progress(r)

	var b [12]byte

	if _, err := iox.ReadAtLeast(r, b[:], len(b)); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("%w: %d-byte input", ErrNotRIFF, len(b))
		}

		return nil, err
	}

	if FourCC(b[0:4]) != RIFF {
		return nil, fmt.Errorf("%w: starts with %q", ErrNotRIFF, b[0:4])
	}

	streams := make(map[FourCC]bool, len(unbounded))
	for _, id := range unbounded {
		streams[id] = true
	}

	return &Reader{
		r:       r,
		size:    binary.LittleEndian.Uint32(b[4:8]),
		form:    FourCC(b[8:12]),
		pos:     idSize,
		streams: streams,
	}, nil
}

// Form returns the form type (e.g. "WAVE").
func (r *Reader) Form() FourCC { return r.form }

// Size returns the declared RIFF size (everything after the size field), which may be 0 or 0xFFFFFFFF for live
// streams.
func (r *Reader) Size() uint32 { return r.size }

// sized reports whether the RIFF size is definite.
func (r *Reader) sized() bool { return r.size != 0 && r.size != UnknownSize }

// declaresMore reports whether a definite RIFF size extends past the current position.
func (r *Reader) declaresMore() bool { return r.sized() && r.pos < int64(r.size) }

// pastEnd reports whether the walk reached the end of a definite RIFF body, which it honours once a stream chunk was
// seen.
func (r *Reader) pastEnd() bool { return r.stream && r.sized() && r.pos >= int64(r.size) }

// Unbounded reports whether the chunk returned by the last call to Next runs to the end of the input: a stream chunk
// whose size is unknown.
func (r *Reader) Unbounded() bool { return r.done }

// Next skips whatever is left of the current chunk (and its pad byte) and returns the next chunk's header together with
// a reader bounded to its body. It returns io.EOF when there are no more chunks. A chunk header cut short by the end of
// the input, or by the end of a definite RIFF body, returns ErrTruncated; a non-printable ID returns ErrBadID. Once the
// RIFF body bounds the walk, a body declared to run past its end is cut there (Header.Size still reports the declared
// size). Any error other than io.EOF is sticky: the walk cannot resume, so every later call returns it.
//
// The returned body reader is valid until the next call to Next.
func (r *Reader) Next() (Header, io.Reader, error) {
	if r.err != nil {
		return Header{}, nil, r.err
	}

	h, body, err := r.next()
	if err != nil && !errors.Is(err, io.EOF) {
		r.err = err
	}

	return h, body, err
}

func (r *Reader) next() (Header, io.Reader, error) {
	if err := r.skip(); err != nil {
		return Header{}, nil, err
	}

	if r.done || r.pastEnd() {
		return Header{}, nil, io.EOF
	}

	// once a stream chunk was seen, a definite RIFF size bounds headers and bodies: nothing is read past it
	bounded := r.stream && r.sized()

	var b [headerSize]byte

	want := len(b)
	if bounded {
		want = int(min(int64(want), int64(r.size)-r.pos))
	}

	n, err := iox.ReadAtLeast(r.r, b[:want], want)
	r.pos += int64(n)

	switch {
	case errors.Is(err, io.EOF) && n == 0:
		return Header{}, nil, io.EOF
	case errors.Is(err, io.EOF), err == nil && want < len(b):
		return Header{}, nil, ErrTruncated
	case err != nil:
		return Header{}, nil, err
	}

	h := Header{ID: FourCC(b[0:4]), Size: binary.LittleEndian.Uint32(b[4:8])}
	if !h.ID.Valid() {
		return h, nil, fmt.Errorf("%w: %q", ErrBadID, b[0:4])
	}

	if r.streams[h.ID] {
		r.stream = true

		// a 0 size is a live-stream marker, unless a definite RIFF size says that more chunks follow
		if !bounded && (h.Size == UnknownSize || (h.Size == 0 && !r.declaresMore())) {
			r.done = true

			return h, r.r, nil
		}
	}

	size := int64(h.Size)
	if bounded {
		size = min(size, int64(r.size)-r.pos)
	}

	r.cur, r.curSize = &io.LimitedReader{R: r.r, N: size}, size
	r.pad = h.Size&1 == 1

	return h, r.cur, nil
}

// skip discards the unread part of the current chunk and its pad byte. A pad byte missing at the end of the input, or
// beyond a definite RIFF size, is not read.
func (r *Reader) skip() error {
	if r.cur == nil {
		return nil
	}

	cur := r.cur
	r.cur = nil

	_, err := io.Copy(io.Discard, cur)
	r.pos += r.curSize - cur.N // what the caller read plus what was discarded

	if err != nil {
		return err
	}

	if r.pad && cur.N == 0 && !r.pastEnd() {
		var b [1]byte

		n, err := iox.ReadAtLeast(r.r, b[:], 1)
		r.pos += int64(n)

		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}

	r.pad = false

	return nil
}

// Writer writes a RIFF file chunk by chunk.
//
// When the destination can be patched (see Close), Close back-patches the RIFF size, the size of a chunk written with
// WriteStream and the bodies of chunks written with Reserve; otherwise those sizes are written as UnknownSize, as live
// streams do, and reserved chunks keep their initial body.
//
// The first write error is sticky: every later call returns it, so a partially written header is never followed by
// more chunks.
type Writer struct {
	w       io.Writer
	written int64 // bytes written after the RIFF size field

	streamAt   int64 // offset of the open stream chunk's size field (counted like written), or -1
	streamSize int64
	reserved   []*Reservation

	err      error // the first write error
	closed   bool
	closeErr error
}

// Reservation is a chunk written with Reserve, whose body Close rewrites in place.
type Reservation struct {
	at   int64 // offset of the body, counted like Writer.written
	body []byte
}

// Set replaces the body that Close writes back. It must have the length of the body passed to Reserve.
func (r *Reservation) Set(body []byte) error {
	if len(body) != len(r.body) {
		return fmt.Errorf("%w: reserved %d bytes, got %d", ErrSize, len(r.body), len(body))
	}

	copy(r.body, body)

	return nil
}

// NewWriter writes the RIFF header for the given form type.
func NewWriter(w io.Writer, form FourCC) (*Writer, error) {
	if w == nil {
		return nil, ErrNilWriter
	}

	rw := &Writer{w: w, streamAt: -1}

	var b [12]byte

	copy(b[0:4], RIFF[:])
	binary.LittleEndian.PutUint32(b[4:8], UnknownSize)
	copy(b[8:12], form[:])

	if err := rw.write(b[:]); err != nil {
		return nil, err
	}

	rw.written = idSize // only the form type counts: the RIFF size covers what follows its own field

	return rw, nil
}

// usable returns why no more chunks can be written, if so.
func (w *Writer) usable() error {
	switch {
	case w.closed || w.streamAt >= 0:
		return ErrClosed
	default:
		return w.err
	}
}

// WriteChunk writes a complete chunk, followed by a pad byte when its size is odd. A body of 4 GiB or more returns
// ErrSize, as its size cannot be written.
func (w *Writer) WriteChunk(id FourCC, body []byte) error {
	if err := w.usable(); err != nil {
		return err
	}

	if !id.Valid() {
		return fmt.Errorf("%w: %q", ErrBadID, id[:])
	}

	if int64(len(body)) > math.MaxUint32 {
		return fmt.Errorf("%w: a %d-byte chunk body", ErrSize, len(body))
	}

	var hdr [headerSize]byte

	copy(hdr[0:4], id[:])
	binary.LittleEndian.PutUint32(hdr[4:8], uint32(len(body))) //nolint:gosec // checked above

	if err := w.write(hdr[:]); err != nil {
		return err
	}

	if err := w.write(body); err != nil {
		return err
	}

	if len(body)%2 == 1 {
		return w.write([]byte{0})
	}

	return nil
}

// Reserve writes a chunk whose body is not final yet, such as WAV's "fact" sample count, and returns a Reservation to
// update it. Close writes the latest body back in place when the destination can be patched.
func (w *Writer) Reserve(id FourCC, body []byte) (*Reservation, error) {
	if err := w.WriteChunk(id, body); err != nil {
		return nil, err
	}

	r := &Reservation{at: w.written - int64(len(body)+len(body)%2), body: bytes.Clone(body)}
	w.reserved = append(w.reserved, r)

	return r, nil
}

// WriteStream starts a final chunk of unknown length (such as WAV's "data" when encoding a live stream) and returns a
// writer for its body. No chunk can follow it. Its size, and the RIFF size, are back-patched by Close when possible.
func (w *Writer) WriteStream(id FourCC) (io.Writer, error) {
	if err := w.usable(); err != nil {
		return nil, err
	}

	if !id.Valid() {
		return nil, fmt.Errorf("%w: %q", ErrBadID, id[:])
	}

	var hdr [headerSize]byte

	copy(hdr[0:4], id[:])
	binary.LittleEndian.PutUint32(hdr[4:8], UnknownSize)

	if err := w.write(hdr[:]); err != nil {
		return nil, err
	}

	w.streamAt = w.written - idSize // the size field is the last 4 bytes of the header just written

	return streamWriter{w}, nil
}

type streamWriter struct{ w *Writer }

func (s streamWriter) Write(p []byte) (int, error) {
	switch {
	case s.w.closed:
		return 0, ErrClosed
	case s.w.err != nil:
		return 0, s.w.err
	}

	n, err := s.w.writeN(p)
	s.w.streamSize += int64(n)

	return n, err
}

func (w *Writer) write(p []byte) error {
	_, err := w.writeN(p)

	return err
}

// writeN writes p to the destination, counting what was written and latching the first error (a short write without
// an error is io.ErrShortWrite).
func (w *Writer) writeN(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	n, err := w.w.Write(p)
	w.written += int64(n)

	if err == nil && n < len(p) {
		err = io.ErrShortWrite
	}

	if err != nil {
		w.err = err
	}

	return n, err
}

// Close finishes the file. It does not close the destination, and calling it again returns the same result.
//
// When the destination can be patched, Close back-patches the stream chunk's size, the RIFF size and reserved chunks,
// and then pads a stream chunk of odd length. Otherwise it writes nothing: a stream chunk of unknown size runs to the
// end of the input, so a pad byte would be read as content. A destination cannot be patched when it does not
// implement io.WriteSeeker, when seeking fails (an *os.File for a pipe), when its offset does not count what was
// written (Linux /dev/null), or when every write goes to its end (a file opened with O_APPEND, including stdout
// redirected with >>). The last case is detected by rewriting the "RIFF" magic in place: if the file grows by those 4
// bytes instead, it is truncated back when it has a Truncate method; otherwise, or on any other change in size, Close
// returns ErrAppendOnly. A stream chunk of 0xFFFFFFFF bytes or more keeps the unknown size and gets no pad byte.
//
// If the stream chunk's size cannot be patched, nothing else is written, so the file still reads as a stream. A write
// error from an earlier call is returned as is.
func (w *Writer) Close() error {
	if w.closed {
		return w.closeErr
	}

	w.closed = true
	w.closeErr = w.finish()

	return w.closeErr
}

func (w *Writer) finish() error {
	if w.err != nil {
		return w.err
	}

	patch, end, err := w.patcher()
	if err != nil || patch == nil {
		return err
	}

	start := end - w.written - headerSize // written counts the bytes after the RIFF size field
	// a stream of 0xFFFFFFFF bytes or more keeps the unknown size, which runs to the end: it gets no pad byte
	pad := w.streamAt >= 0 && w.streamSize%2 == 1 && w.streamSize < UnknownSize
	riffSize := w.written

	if pad {
		riffSize++
	}

	if w.streamAt >= 0 {
		if err := patch(start+headerSize+w.streamAt, le32(w.streamSize)); err != nil {
			return err
		}
	}

	if err := patch(start+idSize, le32(riffSize)); err != nil {
		return err
	}

	for _, r := range w.reserved {
		if err := patch(start+headerSize+r.at, r.body); err != nil {
			return err
		}
	}

	if pad {
		return w.write([]byte{0})
	}

	return nil
}

// le32 encodes a size, clamped to the largest a RIFF field holds.
func le32(v int64) []byte {
	return binary.LittleEndian.AppendUint32(nil, uint32(min(max(v, 0), math.MaxUint32)))
}

// patcher returns a function that overwrites bytes already written and the destination's current offset (the end of
// the RIFF data), or a nil function when the destination cannot be patched (see Close).
func (w *Writer) patcher() (patch func(off int64, b []byte) error, end int64, err error) {
	s, ok := w.w.(io.WriteSeeker)
	if !ok {
		return nil, 0, nil
	}

	end, err = s.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, 0, nil //nolint:nilerr // not seekable after all (a pipe): the file stays a stream
	}

	if end < w.written+headerSize {
		return nil, 0, nil // the offset does not count what was written (Linux /dev/null): the file stays a stream
	}

	patch = func(off int64, b []byte) error {
		if _, err := s.Seek(off, io.SeekStart); err != nil {
			return err
		}

		if _, err := s.Write(b); err != nil {
			return err
		}

		_, err := s.Seek(end, io.SeekStart)

		return err
	}

	inPlace := false

	if wa, ok := w.w.(io.WriterAt); ok {
		if _, err := wa.WriteAt(nil, end); err != nil {
			return nil, 0, nil //nolint:nilerr // an *os.File opened with O_APPEND refuses WriteAt: the file stays a stream
		}

		inPlace = true
		patch = func(off int64, b []byte) error {
			_, err := wa.WriteAt(b, off)

			return err
		}
	}

	return w.probe(s, patch, end, inPlace)
}

// probe rewrites the "RIFF" magic in place and checks that the destination did not grow, as it does when its
// descriptor appends every write. A destination that writes in place through io.WriterAt but cannot seek to its end
// (io.OffsetWriter) skips the check. Only growth of exactly the probe's 4 bytes is undone; any other change in size
// returns ErrAppendOnly without truncating anything.
func (w *Writer) probe(s io.WriteSeeker, patch func(off int64, b []byte) error, end int64, inPlace bool,
) (func(off int64, b []byte) error, int64, error) {
	size, err := s.Seek(0, io.SeekEnd)

	switch {
	case err != nil && inPlace:
		if _, err := s.Seek(end, io.SeekStart); err != nil {
			return nil, 0, err
		}

		return patch, end, nil
	case err != nil:
		return nil, 0, err
	}

	if err := patch(end-w.written-headerSize, RIFF[:]); err != nil {
		return nil, 0, err
	}

	grown, err := s.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, 0, err
	}

	if _, err := s.Seek(end, io.SeekStart); err != nil {
		return nil, 0, err
	}

	if grown == size {
		return patch, end, nil
	}

	if grown != size+idSize { // not (only) the probe's own bytes: never truncate what someone else wrote
		return nil, 0, ErrAppendOnly
	}

	if t, ok := w.w.(interface{ Truncate(size int64) error }); ok && t.Truncate(size) == nil {
		if _, err := s.Seek(size, io.SeekStart); err != nil {
			return nil, 0, err
		}

		return nil, 0, nil // appends every write: the file stays a stream
	}

	return nil, 0, ErrAppendOnly
}
