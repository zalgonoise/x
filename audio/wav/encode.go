package wav

import (
	"bytes"
	"encoding/binary"
)

func (w *Wav) Bytes() ([]byte, error) {
	switch w.Header.BitsPerSample {
	case bitDepth8:
		data := w.genData(1, func(v int, b []byte) []byte {
			return append(b, byte(v))
		})
		return w.bytes(data), nil
	case bitDepth16:
		data := w.genData(2, func(v int, b []byte) []byte {
			return binary.LittleEndian.AppendUint16(b, uint16(v))
		})
		return w.bytes(data), nil

	case bitDepth24:
		data := w.genData(3, func(v int, b []byte) []byte {
			return encode24BitLE(b, int32(v))
		})
		return w.bytes(data), nil

	case bitDepth32:
		data := w.genData(4, func(v int, b []byte) []byte {
			return binary.LittleEndian.AppendUint32(b, uint32(v))
		})
		return w.bytes(data), nil

	default:
		return nil, ErrInvalidBitDepth
	}
}

func (w *Wav) genData(multiplier int, fn func(int, []byte) []byte) []byte {
	data := make([]byte, 0, len(w.Data)*multiplier)
	for i := 0; i < len(w.Data); i++ {
		data = fn(w.Data[i], data)
	}
	return data
}

func (w *Wav) bytes(data []byte) []byte {
	var size uint32 = 4
	var chunkHeaders [][]byte

	for _, chunk := range w.Chunks {
		chunkHeaders = append(chunkHeaders, chunk.Bytes())

		size += 8 + chunk.Subchunk2Size
		switch string(chunk.Subchunk2ID[:]) {
		case junkSubchunkIDString:
			if chunk.Subchunk2Size == 0 {
				chunk.Subchunk2Size = uint32(len(w.Junk))
			}
			chunkHeaders = append(chunkHeaders, w.Junk)

		case dataSubchunkIDString:
			if chunk.Subchunk2Size == 0 {
				chunk.Subchunk2Size = uint32(len(data))
			}
			chunkHeaders = append(chunkHeaders, data)
		}
	}

	if w.Header.ChunkSize == 0 {
		w.Header.ChunkSize = size
	}
	out := bytes.NewBuffer(make([]byte, 0, size+32))
	_, _ = out.Write(w.Header.Bytes())

	for _, chunk := range chunkHeaders {
		_, _ = out.Write(chunk)
	}

	return out.Bytes()
}

func encode24BitLE(buf []byte, v int32) []byte {
	return append(buf, byte(v), byte(v>>8), byte(v>>16))
}
