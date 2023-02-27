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
		offset int
		header *WavHeader
		end    int = headerLen
	)

	if header, err = HeaderFrom(buf[:end]); err == nil {
		offset += end
		w.Header = header
	}
	if err != nil && !errors.Is(err, ErrInvalidHeader) {
		return offset, err
	}
	if w.Header == nil {
		return offset, ErrMissingHeader
	}

	for offset < len(buf) {
		// try to read subchunk headers
		end = 8
		if offset+end < len(buf) {
			if subchunk, err := data.HeaderFrom(buf[offset : offset+end]); err == nil {
				offset += end
				chunk := NewChunk(w.Header.BitsPerSample, subchunk)
				w.Data = chunk

				end = int(subchunk.Subchunk2Size)
				// grab remaining bytes if the byte slice isn't long enough
				// for a subchunk read
				if offset+end+8 > len(buf) {
					end += len(buf) - (offset + end)
				}

				chunk.Parse(buf[offset:offset+end], 0)
				w.Chunks = append(w.Chunks, chunk)
				offset += end
				continue
			}
		}

		if w.Data != nil {
			w.Data.Parse(buf[offset:], offset)
			return len(buf) - offset, nil
		}
		return offset, ErrMissingDataBuffer
	}
	return offset, nil
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
