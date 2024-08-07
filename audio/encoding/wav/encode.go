package wav

import (
	"bytes"

	"github.com/zalgonoise/x/audio/encoding/wav/data"
)

// Read implements the io.Reader interface.
//
// Read will write to the input slice of bytes `buf` the contents
// of the Wav `w`.
//
// It returns the number of bytes written to the buffer, and an error if the buffer
// is not big enough.
func (w *Wav) Read(buf []byte) (n int, err error) {
	if !w.readOnly.Load() {
		w.encode()
		w.readOnly.Store(true)
	}

	return w.buf.Read(buf)
}

// Bytes casts the contents of the Wav `w` as a slice of bytes, with WAV file encoding.
func (w *Wav) Bytes() []byte {
	if !w.readOnly.Load() {
		w.encode()
		w.readOnly.Store(true)
	}

	return w.buf.Bytes()
}

func (w *Wav) encode() {
	var (
		n    int
		size = Size
	)

	for i := range w.Chunks {
		size += data.Size + int(w.Chunks[i].Header().Subchunk2Size)
	}

	if w.Header.ChunkSize == 0 {
		w.Header.ChunkSize = uint32(size)
	}

	buf := make([]byte, size)

	//nolint:errcheck // reading from the header should not raise any errors, and can be safely ignored.
	_, _ = w.Header.Read(buf[n : n+Size])
	n += Size

	for i := range w.Chunks {
		var (
			chunkHeader = w.Chunks[i].Header()
			chunkSize   = int(chunkHeader.Subchunk2Size)
		)

		//nolint:errcheck // reading from the chunk header should not raise any errors, and can be safely ignored.
		_, _ = chunkHeader.Read(buf[n : n+data.Size])
		n += data.Size

		//nolint:errcheck // reading from the chunk should not raise any errors, and can be safely ignored.
		_, _ = w.Chunks[i].Read(buf[n : n+chunkSize])
		n += chunkSize
	}

	w.readOnly.Store(true)
	w.buf = bytes.NewBuffer(buf)
}
