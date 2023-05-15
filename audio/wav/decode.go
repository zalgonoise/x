package wav

import (
	"bytes"
	"errors"

	"github.com/zalgonoise/x/audio/wav/data"
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
	if w.readOnly {
		w.buf.Reset()
		w.readOnly = false
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

// Decode will parse the input slice of bytes `buf` and build a Wav object
// with that data returning a pointer to one, and an error if the buffer is too
// short, or if the data is invalid.
func Decode(buf []byte) (w *Wav, err error) {
	if len(buf) < headerLen {
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
		if w.buf.Len() < headerLen {
			return 0, ErrShortDataBuffer
		}

		var (
			header *WavHeader
			end    = headerLen
		)

		headerBuffer := make([]byte, headerLen)
		if _, err = w.buf.Read(headerBuffer); err != nil {
			return 0, err
		}

		if header, err = HeaderFrom(headerBuffer); err == nil {
			n += end
			w.Header = header
		}

		if err != nil && !errors.Is(err, ErrInvalidHeader) {
			return n, err
		}

		// header is required beyond this point, as w.header.BitsPerSample is necessary
		if w.Header == nil {
			return n, ErrMissingHeader
		}
	}

	for w.buf.Len() > 0 {
		if w.Data != nil {
			end := int(w.Data.Header().Subchunk2Size)
			ln := w.buf.Len()
			if end > 0 && end != ln {
				return n, nil
			}

			chunkBuffer := make([]byte, ln)
			if _, err = w.buf.Read(chunkBuffer); err != nil {
				return 0, err
			}

			w.Data.Parse(chunkBuffer)
			return ln, nil
		}

		// try to read subchunk headers
		end := 8
		if end < w.buf.Len() {
			var (
				subchunk       *data.ChunkHeader
				subchunkBuffer = make([]byte, 8)
			)

			if _, err = w.buf.Read(subchunkBuffer); err != nil {
				return 0, err
			}

			if subchunk, err = data.HeaderFrom(subchunkBuffer); err == nil {
				n += end
				chunk := NewChunk(w.Header.BitsPerSample, subchunk, w.Header.AudioFormat)
				if string(subchunk.Subchunk2ID[:]) == dataSubchunkID {
					w.Data = chunk
				}

				end = int(subchunk.Subchunk2Size)
				// grab remaining bytes if the byte slice isn't long enough
				// for a subchunk read
				if end > 0 && end > w.buf.Len() {
					w.Chunks = append(w.Chunks, chunk)
					return n, nil
				}

				chunkBuffer := make([]byte, end)
				if _, err = w.buf.Read(chunkBuffer); err != nil {
					return 0, err
				}

				chunk.Parse(chunkBuffer)
				w.Chunks = append(w.Chunks, chunk)
				n += end
				continue
			}
		}

		return n, ErrMissingDataBuffer
	}

	return n, nil
}
