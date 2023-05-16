package stream

import (
	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/data"
	"github.com/zalgonoise/x/audio/wav/header"
)

func (w *Wav) parseHeader(buf []byte) error {
	head, err := header.From(buf)
	if err != nil {
		return err
	}
	w.Header = head
	return nil
}

func (w *Wav) parseSubChunk(buf []byte) error {
	subchunk, err := data.HeaderFrom(buf)
	if err != nil {
		return err
	}
	chunk := wav.NewChunk(w.Header.BitsPerSample, subchunk, w.Header.AudioFormat)
	w.Chunks = append(w.Chunks, chunk)
	w.Data = chunk
	return nil
}
