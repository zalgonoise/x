package sdk_test

import (
	"context"
	"sync"

	"github.com/zalgonoise/x/audio"
)

// recorder is an Exporter that keeps a copy of every Buffer. When gate is set, each Export waits for a value from it.
type recorder struct {
	gate      chan struct{}
	exportErr error

	mu        sync.Mutex
	bufs      []audio.Buffer
	flushes   int
	shutdowns int
}

func (r *recorder) Export(ctx context.Context, b audio.Buffer) error {
	if r.gate != nil {
		select {
		case <-r.gate:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.bufs = append(r.bufs, b.Clone())

	return r.exportErr
}

func (r *recorder) ForceFlush(context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.flushes++

	return nil
}

func (r *recorder) Shutdown(context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.shutdowns++

	return nil
}

func (r *recorder) buffers() []audio.Buffer {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]audio.Buffer(nil), r.bufs...)
}

func (r *recorder) counts() (exports, flushes, shutdowns int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.bufs), r.flushes, r.shutdowns
}

// samples concatenates the samples of every recorded Buffer.
func (r *recorder) samples() []float64 {
	var out []float64

	for _, b := range r.buffers() {
		out = append(out, b.Samples...)
	}

	return out
}
