package wav

import (
	"errors"

	"github.com/zalgonoise/x/audio/wav/data"
)

// Write implements the io.Writer interface
//
// Write will gradually build a Wav object from the data passed through the
// slice of bytes `buf` input parameter. This method follows the lifetime of a
// Wav file from start to finish, even if it is raw and without a header.
//
// The method returns the number of bytes read by the buffer, and an error if the
// data is invalid (or too short)
func (w *Wav) Write(buf []byte) (n int, err error) {
	if w.Header == nil && len(buf) < headerLen {
		return 0, ErrShortDataBuffer
	}

	var (
		header *WavHeader
		end    int = headerLen
	)

	if header, err = HeaderFrom(buf[:end]); err == nil {
		n += end
		w.Header = header
	}
	if err != nil && !errors.Is(err, ErrInvalidHeader) {
		return n, err
	}

	// header is required beyond this point, as w.Header.BitsPerSample is necessary
	if w.Header == nil {
		return n, ErrMissingHeader
	}

	for n < len(buf) {
		// try to read subchunk headers
		end = 8
		if n+end < len(buf) {
			if subchunk, err := data.HeaderFrom(buf[n : n+end]); err == nil {
				n += end
				chunk := NewChunk(w.Header.BitsPerSample, subchunk)
				w.Data = chunk

				end = int(subchunk.Subchunk2Size)
				// grab remaining bytes if the byte slice isn't long enough
				// for a subchunk read
				if n+end+8 > len(buf) {
					end += len(buf) - (n + end)
				}

				chunk.Parse(buf[n : n+end])
				w.Chunks = append(w.Chunks, chunk)
				n += end
				continue
			}
		}

		if w.Data != nil {
			w.Data.Parse(buf[n:])
			return len(buf) - n, nil
		}
		return n, ErrMissingDataBuffer
	}
	return n, nil
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
