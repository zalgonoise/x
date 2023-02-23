package wav

import "encoding/binary"

func Decode(buf []byte) (*Wav, error) {
	if len(buf) < 44 {
		return nil, ErrShortBuffer
	}
	wav := new(Wav)

	var offset = 44
	header, err := HeaderFrom(buf[:offset])
	if err != nil {
		return nil, err
	}
	wav.Header = header

	if string(header.Subchunk2ID[:]) == "junk" {
		wav.Junk = buf[offset : offset+int(header.Subchunk2Size)]
		offset += int(header.Subchunk2Size)
		data, err := SubChunkFrom(buf[offset : offset+8])
		if err != nil {
			return nil, err
		}
		wav.Chunks = append(wav.Chunks,
			&SubChunk{
				Subchunk2ID:   junkSubchunk2ID,
				Subchunk2Size: header.Subchunk2Size,
			},
			data,
		)
		offset += 8
	}

	err = wav.parseData(buf[offset:])
	if err != nil {
		return nil, err
	}
	return wav, nil
}

func (w *Wav) parseData(buf []byte) error {
	switch w.Header.BitsPerSample {
	case bitDepth8:
		w.Data = make([]int, len(buf))
		for i := 0; i < len(buf); i++ {
			w.Data[i] = int(uint8(buf[i]))
		}
		return nil
	case bitDepth16:
		w.Data = make([]int, len(buf)/2)
		for i, j := 0, 0; i+1 < len(buf); i, j = i+2, j+1 {
			w.Data[j] = int(int16(binary.LittleEndian.Uint16(buf[i : i+2])))
		}
		return nil

	case bitDepth24:
		w.Data = make([]int, len(buf)/3)
		for i, j := 0, 0; i+2 < len(buf); i, j = i+3, j+1 {
			w.Data[j] = int(int32(decode24BitLE(buf[i : i+3])))
		}
		return nil

	case bitDepth32:
		w.Data = make([]int, len(buf)/4)
		for i, j := 0, 0; i+3 < len(buf); i, j = i+4, j+1 {
			w.Data[j] = int(int32(binary.LittleEndian.Uint32(buf[i : i+4])))
		}
		return nil

	default:
		return ErrInvalidBitDepth
	}
}

func decode24BitLE(buf []byte) int32 {
	value := int32(buf[0]) | (int32(buf[1]) << 8) | (int32(buf[2]) << 16)
	if value&0x00800000 != 0 {
		value |= -16777216 // handle signed integers
	}
	return value
}
