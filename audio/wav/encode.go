package wav

import "fmt"

func (w *Wav) Read(buf []byte) (n int, err error) {
	size, data := w.encode()
	if len(buf) < size {
		return n, fmt.Errorf("%w: input buffer with length %d cannot fit %d bytes", ErrShortDataBuffer, len(buf), size)
	}

	for i := range data {
		n += copy(buf[n:], data[i])
	}
	return size, nil
}

func (w *Wav) Bytes() ([]byte, error) {
	var n int
	size, data := w.encode()

	buf := make([]byte, size+32)
	for i := range data {
		n += copy(buf[n:], data[i])
	}
	return buf, nil
}

func (w *Wav) encode() (size int, data [][]byte) {
	size = 4
	data = make([][]byte, (len(w.Chunks)*2)+1)

	for i, j := 0, 1; i < len(w.Chunks); i, j = i+1, j+2 {
		data[j] = w.Chunks[i].Header().Bytes()
		data[j+1] = w.Chunks[i].Generate()
		size += 8 + len(data[j+1])
	}
	if w.Header.ChunkSize == 0 {
		w.Header.ChunkSize = uint32(size)
	}
	data[0] = w.Header.Bytes()
	return size, data
}
