package wav

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/zalgonoise/gbuf"
	"github.com/zalgonoise/x/audio/wav/data"
)

const (
	minBufferSize = 16
)

type WavBuffer struct {
	Header  *WavHeader
	Chunks  []data.Chunk
	Data    data.Chunk
	Filters []StreamFilter
	Reader  io.Reader
	ring    *gbuf.RingFilter[byte]
	ratio   float64
}

func (w *WavBuffer) Stream(ctx context.Context, errCh chan<- error) {
	err := w.stream(ctx)
	if err != nil {
		errCh <- err
		return
	}
}

func NewStream(r io.Reader) *WavBuffer {
	return &WavBuffer{
		Reader: r,
		ratio:  1.0,
	}
}

func (w *WavBuffer) WithFilter(fns ...StreamFilter) {
	for _, fn := range fns {
		if fn != nil {
			w.Filters = append(w.Filters, fn)
		}
	}
}

func (w *WavBuffer) Ratio(ratio float64) {
	if ratio == 0 {
		return
	}
	w.ratio = ratio
}

func (w *WavBuffer) parseHeader(buf []byte) error {
	header, err := HeaderFrom(buf)
	if err != nil {
		return err
	}
	w.Header = header
	return nil
}

func (w *WavBuffer) parseSubChunk(buf []byte) error {
	subchunk, err := data.HeaderFrom(buf)
	if err != nil {
		return err
	}
	chunk := NewChunk(w.Header.BitsPerSample, subchunk)
	w.Chunks = append(w.Chunks, chunk)
	w.Data = chunk
	return nil
}

func (w *WavBuffer) processChunk(buf []byte) error {
	w.Data.Parse(buf)
	if len(w.Filters) == 0 {
		return nil
	}
	v := w.Data.Value()
	defer w.Data.Reset()
	for _, fn := range w.Filters {
		err := fn(w, v, buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *WavBuffer) stream(ctx context.Context) error {
	hbuf := make([]byte, 36)
	if _, err := w.Reader.Read(hbuf); err != nil {
		return err
	}
	if err := w.parseHeader(hbuf); err != nil && w.Header == nil {
		return err
	}
	bufferSize := int(w.Header.ByteRate)
	if float64(bufferSize)*w.ratio >= minBufferSize {
		bufferSize = int(float64(bufferSize) * w.ratio)
	}
	w.ring = gbuf.NewRingFilter(bufferSize, w.processChunk)
	scbuf := make([]byte, 8)
	if _, err := w.Reader.Read(scbuf); err != nil {
		return err
	}
	if err := w.parseSubChunk(scbuf); err != nil {
		return err
	}

	var err error
	go func() {
		if _, err = w.ring.ReadFrom(w.Reader); err != nil {
			return
		}
	}()

	for {
		select {
		case <-ctx.Done():
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			return nil
		default:
		}
	}
}

func (w *WavBuffer) Bytes() ([]byte, error) {
	var n int
	size, data := w.encode()

	buf := make([]byte, size+32)
	for i := range data {
		n += copy(buf[n:], data[i])
	}
	return buf, nil
}

func (w *WavBuffer) encode() (size int, byteData [][]byte) {
	size = 4
	byteData = make([][]byte, 3)

	for i, j := 0, 1; i < len(w.Chunks); i, j = i+1, j+2 {
		byteData[j] = w.Chunks[i].Header().Bytes()
		byteData[j+1] = w.Chunks[i].Generate()
		size += 8 + len(byteData[j+1])
	}

	if w.Header.ChunkSize == 0 {
		w.Header.ChunkSize = uint32(size)
	}
	byteData[0] = w.Header.Bytes()
	return size, byteData
}

func (w *WavBuffer) Read(buf []byte) (n int, err error) {
	size, data := w.encode()
	if len(buf) < size {
		return n, fmt.Errorf("%w: input buffer with length %d cannot fit %d bytes", ErrShortDataBuffer, len(buf), size)
	}

	for i := range data {
		n += copy(buf[n:], data[i])
	}
	return size, nil
}
