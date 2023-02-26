package wav

func (w *Wav) Read(buf []byte) (n int, err error) {
	var size int = 4
	var data = make([][]byte, (len(w.Chunks)*2)+1)

	for i, j := 0, 1; i < len(w.Chunks); i, j = i+1, j+2 {
		header := w.Chunks[i].Header()
		data[j] = header.Bytes()
		data[j+1] = w.Chunks[i].Generate()
		size += 8 + len(data[j+1])
	}
	if w.Header.ChunkSize == 0 {
		w.Header.ChunkSize = uint32(size)
	}
	data[0] = w.Header.Bytes()

	for i := range data {
		m := copy(buf[n:], data[i])
		n += m
		if m < len(data[i]) {
			return n, ErrShortDataBuffer
		}
	}
	return n, nil
}

func (w *Wav) Bytes() ([]byte, error) {
	var n int
	var size int = 4
	var data = make([][]byte, (len(w.Chunks)*2)+1)

	for i, j := 0, 1; i < len(w.Chunks); i, j = i+1, j+2 {
		data[j] = w.Chunks[i].Header().Bytes()
		data[j+1] = w.Chunks[i].Generate()
		size += 8 + len(data[j+1])
	}
	if w.Header.ChunkSize == 0 {
		w.Header.ChunkSize = uint32(size)
	}
	data[0] = w.Header.Bytes()

	buf := make([]byte, size+32)
	for i := range data {
		n += copy(buf[n:], data[i])
		if n < len(data[i]) {
			return nil, ErrShortDataBuffer
		}
	}
	return buf, nil
}

func encode24BitLE(buf []byte, v int32) []byte {
	return append(buf, byte(v), byte(v>>8), byte(v>>16))
}
