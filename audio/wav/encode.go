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
	var size = 36 + w.Header.Subchunk2Size
	var chunkHeaders [][]byte

	if len(w.Chunks) > 1 {
		// append junk data
		chunkHeaders = append(chunkHeaders, w.Junk)
		for i := 1; i < len(w.Chunks); i++ {
			size += w.Chunks[i].Subchunk2Size
			chunkHeaders = append(chunkHeaders, w.Chunks[i].Bytes())
		}
	}

	out := bytes.NewBuffer(make([]byte, 0, size))
	_, _ = out.Write(w.Header.Bytes())

	if len(chunkHeaders) > 0 {
		for i := 0; i < len(chunkHeaders); i++ {
			_, _ = out.Write(chunkHeaders[i])
		}
	}
	_, _ = out.Write(data)
	return out.Bytes()
}

func encode24BitLE(buf []byte, v int32) []byte {
	return append(buf, byte(v), byte(v>>8), byte(v>>16))
}
