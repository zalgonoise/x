package wav

import (
	"encoding/binary"
)

func (w *Wav) Write(buf []byte) (n int, err error) {
	var offset int
	if w.Header == nil {
		offset += headerLen
		header, err := HeaderFrom(buf[:offset])
		if err != nil {
			return offset, err
		}
		w.Header = header
	}
	if len(w.Chunks) == 0 {
		// data markers
		var start, end int
		for offset < len(buf)-1 {
			data, err := SubChunkFrom(buf[offset : offset+8])
			if err != nil {
				return offset, err
			}
			offset += 8
			w.Chunks = append(w.Chunks, data)
			switch string(data.Subchunk2ID[:]) {
			case dataSubchunkIDString:
				start = offset
				end = offset + int(data.Subchunk2Size)
				if end > len(buf) {
					end = len(buf)
				}
			case junkSubchunkIDString:
				w.Junk = buf[offset : offset+int(data.Subchunk2Size)]
			}
			offset += int(data.Subchunk2Size)
		}

		err = w.parseData(buf[start:end])
		if err != nil {
			return offset, err
		}
		return end, nil
	}
	err = w.parseData(buf)
	if err != nil {
		return offset, err
	}
	return len(buf), nil
}

func Decode(buf []byte) (*Wav, error) {
	if len(buf) < headerLen {
		return nil, ErrShortDataBuffer
	}
	wav := new(Wav)

	var offset = headerLen
	header, err := HeaderFrom(buf[:offset])
	if err != nil {
		return nil, err
	}
	wav.Header = header

	// data markers
	var start, end int

	for offset < len(buf)-1 {
		data, err := SubChunkFrom(buf[offset : offset+8])
		if err != nil {
			return nil, err
		}
		offset += 8
		wav.Chunks = append(wav.Chunks, data)
		switch string(data.Subchunk2ID[:]) {
		case dataSubchunkIDString:
			start = offset
			end = offset + int(data.Subchunk2Size)
			if len(buf) > end {
				end = len(buf)
			}
		case junkSubchunkIDString:
			wav.Junk = buf[offset : offset+int(data.Subchunk2Size)]
		}
		offset += int(data.Subchunk2Size)
	}

	err = wav.parseData(buf[start:end])
	if err != nil {
		return nil, err
	}
	return wav, nil
}

func (w *Wav) parseData(buf []byte) error {
	// var offset int
	switch w.Header.BitsPerSample {
	case bitDepth8:
		data := make([]int, len(buf))
		for i := 0; i < len(buf); i++ {
			data[i] = int(int8(buf[i]))
		}
		w.Data = append(w.Data, data...)
		return nil
	case bitDepth16:
		data := make([]int, len(buf)/2)
		for i, j := 0, 0; i+1 < len(buf); i, j = i+2, j+1 {
			data[j] = int(int16(binary.LittleEndian.Uint16(buf[i : i+2])))
		}
		w.Data = append(w.Data, data...)
		return nil

	case bitDepth24:
		data := make([]int, len(buf)/3)
		for i, j := 0, 0; i+2 < len(buf); i, j = i+3, j+1 {
			data[j] = int(int32(decode24BitLE(buf[i : i+3])))
		}
		w.Data = append(w.Data, data...)
		return nil

	case bitDepth32:
		data := make([]int, len(buf)/4)
		for i, j := 0, 0; i+3 < len(buf); i, j = i+4, j+1 {
			data[j] = int(int32(binary.LittleEndian.Uint32(buf[i : i+4])))
		}
		w.Data = append(w.Data, data...)
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
