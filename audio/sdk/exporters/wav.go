// Package exporters provides sdk.Exporter implementations.
package exporters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/sdk"
)

// ErrFormatChanged is returned by WAV.Export for a Buffer whose format differs from the first one's.
var ErrFormatChanged = errors.New("exporters: audio format changed mid-stream")

// WAV writes every Buffer it receives to an io.Writer as one WAV stream, with the format of the first Buffer. Sizes are
// back-patched on Shutdown when the writer can seek (see wav.Encoder).
type WAV struct {
	w    io.Writer
	opts []wav.EncoderOption

	mu     sync.Mutex
	enc    *wav.Encoder
	closed bool
}

// NewWAV returns an Exporter writing to w. It does not close w.
func NewWAV(w io.Writer, opts ...wav.EncoderOption) *WAV { return &WAV{w: w, opts: opts} }

// Export implements sdk.Exporter.
func (e *WAV) Export(_ context.Context, b audio.Buffer) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return sdk.ErrShutdown
	}

	if e.enc == nil {
		enc, err := wav.NewEncoder(e.w, b.Format, e.opts...)
		if err != nil {
			return err
		}

		e.enc = enc
	}

	if b.Format != e.enc.Format() {
		return fmt.Errorf("%w: %s, then %s", ErrFormatChanged, e.enc.Format(), b.Format)
	}

	_, err := e.enc.WriteSamples(b.Samples)

	return err
}

// ForceFlush implements sdk.Exporter. The encoder writes through, so there is nothing to flush.
func (e *WAV) ForceFlush(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return sdk.ErrShutdown
	}

	return nil
}

// Shutdown implements sdk.Exporter: it finishes the WAV stream (see wav.Encoder.Close). It is idempotent.
func (e *WAV) Shutdown(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil
	}

	e.closed = true

	if e.enc == nil {
		return nil
	}

	return e.enc.Close()
}
