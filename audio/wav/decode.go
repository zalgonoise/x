package wav

import (
	"errors"

	"github.com/zalgonoise/x/audio/wav/data"
)

func (w *Wav) Write(buf []byte) (n int, err error) {
	if len(buf) < headerLen {
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
