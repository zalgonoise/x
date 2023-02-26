package wav

import (
	"errors"
	"io"
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
			if subchunk, err := SubChunkFrom(buf[offset : offset+end]); err == nil {
				offset += end
				chunk := NewDataChunk(w.Header.BitsPerSample, subchunk)
				w.Data = chunk

				end = int(subchunk.Subchunk2Size)
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
		return offset, err

	}
	return offset, nil
}

func Decode(buf []byte) (w *Wav, err error) {
	if len(buf) < headerLen {
		return nil, ErrShortDataBuffer
	}
	w = new(Wav)

	var (
		offset int
		header *WavHeader
	)

	if header, err = HeaderFrom(buf[:headerLen]); err != nil && w.Header == nil {
		return nil, err
	}
	offset += headerLen
	w.Header = header

	for offset+8 < len(buf) {
		var (
			subchunk *SubChunk
			err      error
		)
		if subchunk, err = SubChunkFrom(buf[offset : offset+8]); err != nil {
			if errors.Is(err, io.EOF) {
				return w, nil
			}
			return w, err
		}
		offset += 8
		if offset+int(subchunk.Subchunk2Size) > len(buf) {
			return w, nil
		}
		chunk := NewDataChunk(w.Header.BitsPerSample, subchunk)
		if offset+int(subchunk.Subchunk2Size)+8 > len(buf) {
			chunk.Parse(buf[offset:], 0)
		} else {
			chunk.Parse(buf[offset:offset+int(subchunk.Subchunk2Size)], 0)
		}

		w.Chunks = append(w.Chunks, chunk)
		offset += int(subchunk.Subchunk2Size)
	}
	return w, nil
}

func decode24BitLE(buf []byte) int32 {
	value := int32(buf[0]) | (int32(buf[1]) << 8) | (int32(buf[2]) << 16)
	if value&0x00800000 != 0 {
		value |= -16777216 // handle signed integers
	}
	return value
}
