package wav

import (
	"bytes"

	datah "github.com/zalgonoise/x/audio/wav/data/header"
	"github.com/zalgonoise/x/audio/wav/header"
)

// Read implements the io.Reader interface
//
// Read will write to the input slice of bytes `buf` the contents
// of the Wav `w`.
//
// It returns the number of bytes written to the buffer, and an error if the buffer
// is not big enough
func (w *Wav) Read(buf []byte) (n int, err error) {
	if !w.readOnly.Load() {
		w.encode()
	}

	return w.buf.Read(buf)
}

// Bytes casts the contents of the Wav `w` as a slice of bytes, with WAV file encoding
func (w *Wav) Bytes() []byte {
	if !w.readOnly.Load() {
		w.encode()
	}

	return w.buf.Bytes()
}

func (w *Wav) encode() {
	var (
		n    int
		size = header.Size
	)

	for i := range w.Chunks {
		size += datah.Size + int(w.Chunks[i].Header().Subchunk2Size)
	}

	if w.Header.ChunkSize == 0 {
		w.Header.ChunkSize = uint32(size)
	}

	buf := make([]byte, size)
	_, _ = w.Header.Read(buf[n : n+header.Size])
	n += header.Size

	for i := range w.Chunks {
		var (
			chunkHeader = w.Chunks[i].Header()
			chunkSize   = int(chunkHeader.Subchunk2Size)
		)

		_, _ = chunkHeader.Read(buf[n : n+datah.Size])
		n += datah.Size
		_, _ = w.Chunks[i].Read(buf[n : n+chunkSize])
		n += chunkSize
	}

	w.readOnly.Store(true)
	w.buf = bytes.NewBuffer(buf)

	return
}
