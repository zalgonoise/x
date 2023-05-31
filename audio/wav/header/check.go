package header

import "bytes"

// Check confirms whether the input bytes are likely to be a header
// with the least operations possible
func Check(buf []byte) bool {
	if len(buf) < Size {
		return false
	}

	return bytes.Equal(buf[:chunkIDEnd], defaultChunkID[:]) &&
		bytes.Equal(buf[formatOffset:subChunkIDEnd], formatAndSubchunkID)
}
