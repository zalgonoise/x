package stream

import (
	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/data"
)

func (w *Wav) parseHeader(buf []byte) error {
	header, err := wav.HeaderFrom(buf)
	if err != nil {
		return err
	}
	w.Header = header
	return nil
}

func (w *Wav) parseSubChunk(buf []byte) error {
	subchunk, err := data.HeaderFrom(buf)
	if err != nil {
		return err
	}
	chunk := wav.NewChunk(w.Header.BitsPerSample, subchunk)
	w.Chunks = append(w.Chunks, chunk)
	w.Data = chunk
	return nil
}