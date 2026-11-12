package wav

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/zalgonoise/x/audio/encoding/wav/data"
)

const dataSubchunkID = "data"

// Write implements the io.Writer interface.
//
// Write will gradually build a Wav object from the data passed through the
// slice of bytes `buf` input parameter. This method follows the lifetime of a
// Wav file from start to finish, even if it is raw and without a header.
//
// The method returns the number of bytes read by the buffer, and an error if the
// data is invalid (or too short).
func (w *Wav) Write(buf []byte) (n int, err error) {
	return w.write(buf, w.decode)
}

// write buffers a copy of buf (the caller may reuse it) and runs the given decoder over everything buffered so far.
//
// ErrShortDataBuffer means "more data needed"; any other error is permanent and returned again by later Writes,
// which then stop buffering input. Bytes that follow the RIFF content (padding, appended tags) are ignored.
func (w *Wav) write(buf []byte, decode func() (int, error)) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	// the cached encoding (Read / Bytes) no longer reflects the data
	w.readOnly.Store(false)

	if w.trailing {
		return len(buf), nil
	}

	if w.buf == nil {
		w.buf = new(bytes.Buffer)
	}

	w.buf.Write(buf)

	n, err := decode()
	if err != nil && !errors.Is(err, ErrShortDataBuffer) {
		w.err = err
	}

	return n, err
}

// ReadFrom implements the io.ReaderFrom interface
//
// # It allows for a Wav file (or stream) to be read and decoded into a data structure
//
// This implementation differs from a stream implementation of the Wav data structure, which
// would scope the stored PCM data in a ring buffer, both to save on storage / memory, and
// to only keep the last X bits of an audio stream (usually for analysis).
func (w *Wav) ReadFrom(r io.Reader) (n int64, err error) {
	header := new(Header)

	num, err := header.ReadFrom(r)
	n += num

	if err != nil {
		// never keep a half-decoded header: later calls would trust it
		return n, err
	}

	w.Header = header

	for w.Data == nil {
		h := new(data.Header)

		num, err = h.ReadFrom(r)
		n += num

		if err != nil {
			return n, err
		}

		chunk := NewChunk(h, w.Header.BitsPerSample, w.Header.AudioFormat)
		if chunk == nil {
			return n, fmt.Errorf("%w: BitsPerSample %d", ErrInvalidBitDepth, w.Header.BitsPerSample)
		}

		alignFrames(chunk, w.Header.BlockAlign)
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

// Decode will parse the input slice of bytes `buf` and build a Wav object
// with that data returning a pointer to one, and an error if the buffer is too
// short, or if the data is invalid.
func Decode(buf []byte) (w *Wav, err error) {
	if len(buf) < Size {
		return nil, ErrShortDataBuffer
	}

	w = new(Wav)

	_, err = w.Write(buf)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Wav) decode() (n int, err error) {
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

	return w.decodeChunks(n, func(h *data.Header) Chunk {
		return NewChunk(h, w.Header.BitsPerSample, w.Header.AudioFormat)
	})
}

// decodeHeader consumes the RIFF/fmt header from the Write buffer, once all of it (including any fmt extension
// bytes) is buffered; until then it consumes nothing and returns ErrShortDataBuffer.
func (w *Wav) decodeHeader() (n int, err error) {
	if w.buf.Len() < Size {
		return 0, ErrShortDataBuffer
	}

	// validate the fixed part first: input that is not a WAV header must fail now, not wait for an "extension"
	if _, err := HeaderFrom(w.buf.Bytes()[:Size]); err != nil {
		return 0, err
	}

	if ext := fmtExtension(w.buf.Bytes()); ext > 0 && int64(w.buf.Len()) < Size+ext {
		return 0, ErrShortDataBuffer
	}

	header := new(Header)

	num, err := header.ReadFrom(w.buf)
	if err != nil {
		// never keep a half-decoded header: the next Write would trust it
		return int(num), err
	}

	w.Header = header

	return int(num), nil
}

// decodeChunks decodes as many sub-chunks from the Write buffer as are complete. Non-data chunks are only consumed
// once their whole body (and pad byte) is buffered; the data chunk is decoded incrementally, in whole frames, and
// bounded by its declared size.
func (w *Wav) decodeChunks(n int, newChunk func(*data.Header) Chunk) (int, error) {
	for w.buf.Len() > 0 {
		if w.padLeft {
			_ = w.buf.Next(1)
			w.padLeft = false
			n++

			continue
		}

		if w.Data != nil && !w.dataDone {
			num := w.decodeData()
			n += num

			if !w.dataDone {
				return n, nil // waiting for more data
			}

			continue
		}

		num, ok, err := w.decodeSubChunk(newChunk)
		n += num

		if err != nil || !ok {
			return n, err
		}
	}

	return n, nil
}

// decodeSubChunk decodes the next sub-chunk header and, for anything but the (first) data chunk, the chunk's whole body.
// ok is false when more input is needed; nothing is consumed in that case.
func (w *Wav) decodeSubChunk(newChunk func(*data.Header) Chunk) (n int, ok bool, err error) {
	if w.buf.Len() < data.Size {
		return 0, false, nil
	}

	h, err := data.From(w.buf.Bytes()[:data.Size])
	if err != nil {
		if w.dataDone {
			// not a chunk after the audio (zero padding, an appended ID3 tag): the RIFF content ended
			n = w.buf.Len()
			w.buf.Reset()
			w.trailing = true

			return n, false, nil
		}

		return 0, false, err
	}

	if w.Data != nil || string(h.Subchunk2ID[:]) != dataSubchunkID {
		// opaque chunk (LIST, fact, JUNK, a second data chunk, ...): keep it whole, with its pad byte
		size := int(h.Subchunk2Size)
		total := data.Size + size + size&1

		if w.buf.Len() < total {
			return 0, false, nil
		}

		_ = w.buf.Next(data.Size)
		body := bytes.Clone(w.buf.Next(size))
		_ = w.buf.Next(size & 1)

		chunk := data.NewJunk(h)
		chunk.Parse(body)
		w.Chunks = append(w.Chunks, chunk)

		return total, true, nil
	}

	_ = w.buf.Next(data.Size)

	chunk := newChunk(h)
	if chunk == nil {
		return data.Size, false, fmt.Errorf("%w: BitsPerSample %d", ErrInvalidBitDepth, w.Header.BitsPerSample)
	}

	w.Chunks = append(w.Chunks, chunk)
	w.Data = chunk

	if h.Subchunk2Size != 0 && h.Subchunk2Size != math.MaxUint32 {
		w.dataBounded = true
		w.dataLeft = int64(h.Subchunk2Size)
		w.dataOdd = h.Subchunk2Size&1 == 1
	}

	return data.Size, true, nil
}

// decodeData decodes buffered bytes of the data chunk, in whole frames, up to the chunk's declared size.
func (w *Wav) decodeData() (n int) {
	var (
		align = max(int(w.Header.BlockAlign), 1)
		ln    = w.buf.Len()
		final bool
	)

	// a data chunk with no declared size (live streams, or one created by New / FromHeader / From) is unbounded
	if w.dataBounded && int64(ln) >= w.dataLeft {
		ln, final = int(w.dataLeft), true
	}

	// only whole frames are decoded: an incomplete frame stays buffered for the next Write, unless the chunk ends there
	whole := ln - ln%align
	if whole > 0 {
		w.Data.Parse(w.buf.Next(whole))
	}

	if !final {
		if w.dataBounded {
			w.dataLeft -= int64(whole)
		}

		return whole
	}

	// a trailing partial frame at the end of a bounded chunk cannot be decoded: drop it
	_ = w.buf.Next(ln - whole)

	w.padLeft = w.dataOdd
	w.dataLeft = 0
	w.dataDone = true

	return ln
}

// fmtExtension returns how many bytes follow the 16-byte fmt chunk body in a buffered header (cbSize and extensible
// fields, plus a pad byte), or 0 when the buffer is too short to tell.
func fmtExtension(b []byte) int64 {
	if len(b) < Size {
		return 0
	}

	size := int64(binary.LittleEndian.Uint32(b[16:20]))
	if size <= defaultSubChunk1Size {
		return 0
	}

	return size - defaultSubChunk1Size + size&1
}
