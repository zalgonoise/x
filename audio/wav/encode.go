package wav

import (
	"bytes"
)

// Read implements the io.Reader interface
//
// Read will write to the input slice of bytes `buf` the contents
// of the Wav `w`.
//
// It returns the number of bytes written to the buffer, and an error if the buffer
// is not big enough
func (w *Wav) Read(buf []byte) (n int, err error) {
	if !w.readOnly {
		if err = w.encode(); err != nil {
			return 0, err
		}
		w.readOnly = true
	}

	if w.buf == nil || w.buf.Len() == 0 {
		w.readOnly = false
		return w.Read(buf)
	}

	return w.buf.Read(buf)
}

// Bytes casts the contents of the Wav `w` as a slice of bytes, with WAV file encoding
func (w *Wav) Bytes() []byte {
	if !w.readOnly {
		if err := w.encode(); err != nil {
			return nil
		}
		w.readOnly = true
	}

	return w.buf.Bytes()
}

func (w *Wav) encode() error {
	size := 4
	data := make([][]byte, (len(w.Chunks)*2)+1)

	for i, j := 0, 1; i < len(w.Chunks); i, j = i+1, j+2 {
		// processing chunks before headers so sizes can be updated if required
		data[j+1] = w.Chunks[i].Bytes()
		data[j] = w.Chunks[i].Header().Bytes()
		size += 8 + len(data[j+1])
	}
	if w.Header.ChunkSize == 0 {
		w.Header.ChunkSize = uint32(size)
	}
	data[0] = w.Header.Bytes()

	w.buf = bytes.NewBuffer(make([]byte, 0, size))

	for i := range data {
		if _, err := w.buf.Write(data[i]); err != nil {
			return err
		}
	}

	return nil
}
