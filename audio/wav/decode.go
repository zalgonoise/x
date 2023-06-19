package wav

import (
	"bytes"
	"io"

	dataheader "github.com/zalgonoise/x/audio/wav/data/header"
	"github.com/zalgonoise/x/audio/wav/header"
)

const dataSubchunkID = "data"

// Write implements the io.Writer interface
//
// Write will gradually build a Wav object from the data passed through the
// slice of bytes `buf` input parameter. This method follows the lifetime of a
// Wav file from start to finish, even if it is raw and without a header.
//
// The method returns the number of bytes read by the buffer, and an error if the
// data is invalid (or too short)
func (w *Wav) Write(buf []byte) (n int, err error) {
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
func (w *Wav) ReadFrom(r io.Reader) (n int64, err error) {
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

		chunk := NewChunk(h, w.Header.BitsPerSample, w.Header.AudioFormat)
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
	if len(buf) < header.Size {
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

func (w *Wav) decodeHeader() (n int, err error) {
	if w.buf.Len() < header.Size {
		return 0, ErrShortDataBuffer
	}

	var head *header.Header

	headerBuffer := make([]byte, header.Size)
	if _, err = w.buf.Read(headerBuffer); err != nil {
		return 0, err
	}

	head, err = header.From(headerBuffer)
	if err != nil {
		return header.Size, err
	}

	w.Header = head
	return header.Size, nil
}

func (w *Wav) decodeNewSubChunk(n int) (int, error) {
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
			chunk := NewChunk(subchunk, w.Header.BitsPerSample, w.Header.AudioFormat)
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

func (w *Wav) decodeIntoData(n int) (int, error) {
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
