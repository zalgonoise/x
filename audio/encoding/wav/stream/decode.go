package stream

import (
	"github.com/zalgonoise/x/audio/encoding/wav"
	datah "github.com/zalgonoise/x/audio/encoding/wav/data/header"
	"github.com/zalgonoise/x/audio/encoding/wav/header"
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
	h, err := datah.From(buf)
	if err != nil {
		return err
	}
	chunk := wav.NewChunk(h, w.Header.BitsPerSample, w.Header.AudioFormat)
	w.Chunks = append(w.Chunks, chunk)
	w.Data = chunk
	return nil
}
