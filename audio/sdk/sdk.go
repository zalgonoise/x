// Package sdk runs audio pipelines: a Source of encoded audio is decoded into samples, optionally transformed, and
// handed to Exporters as audio.Buffer values.
//
// It follows the conventions of the OpenTelemetry SDKs (CLAUDE.md §3, §5): components have an idempotent
// Shutdown(ctx) and ForceFlush(ctx), calls after Shutdown return ErrShutdown instead of panicking, asynchronous errors go
// to a single ErrorHandler, and configuration uses functional options whose zero values are no-ops. Library code never
// installs signal handlers or global state; the caller owns the lifecycle.
//
// This package replaces the sdk/audio Processor/Consumer/Exporter design (Phase 3 of the roadmap); the analysis plane
// (meters, aggregations, readers) builds on it.
package sdk

import (
	"context"
	"errors"
	"io"

	"github.com/zalgonoise/x/audio"
	_ "github.com/zalgonoise/x/audio/encoding/wav" // registers "wav" with audio.Decode, the default decoder
)

// Errors returned by the package.
var (
	// ErrShutdown is returned by calls made after Shutdown, and by a Run that Shutdown stopped.
	ErrShutdown = errors.New("sdk: shut down")
	// ErrRunning is returned by Run while another Run of the same Pipeline is in progress.
	ErrRunning = errors.New("sdk: pipeline already running")
	// ErrNilSource is returned by NewPipeline without a Source.
	ErrNilSource = errors.New("sdk: nil source")
	// ErrNilExporter is returned by NewPipeline when an Exporter is nil.
	ErrNilExporter = errors.New("sdk: nil exporter")
	// ErrInvalidOption is returned for an option value that cannot work (a negative duration, an empty queue...).
	ErrInvalidOption = errors.New("sdk: invalid option")
)

// Exporter receives decoded audio.
//
// Export is called with a Buffer that is valid only for the duration of the call: an Exporter must not modify its
// samples, and must Clone it to keep it. Export may block; wrap a slow Exporter in a QueuedExporter to keep a live
// pipeline real-time. ForceFlush delivers whatever the Exporter buffers; Shutdown flushes and releases it, is
// idempotent, and makes later calls return ErrShutdown.
type Exporter interface {
	Export(ctx context.Context, b audio.Buffer) error
	ForceFlush(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// Source opens a stream of encoded audio, such as a WAV file, an HTTP response body or stdin. Each Run of a Pipeline
// opens it once and closes what it returns. Closing the stream must unblock a pending Read, which is how a Pipeline
// stops a source that is waiting for data.
type Source interface {
	Open(ctx context.Context) (io.ReadCloser, error)
}

// SourceFunc adapts a function to the Source interface.
type SourceFunc func(ctx context.Context) (io.ReadCloser, error)

// Open implements Source.
func (f SourceFunc) Open(ctx context.Context) (io.ReadCloser, error) { return f(ctx) }

// Transform maps each Buffer to another one before it is exported: gain, filters, mixdown, resampling. It must not
// modify its input's samples; the Buffer it returns may use scratch space it owns, valid until its next call.
type Transform interface {
	Transform(ctx context.Context, b audio.Buffer) (audio.Buffer, error)
}

// TransformFunc adapts a function to the Transform interface.
type TransformFunc func(ctx context.Context, b audio.Buffer) (audio.Buffer, error)

// Transform implements Transform.
func (f TransformFunc) Transform(ctx context.Context, b audio.Buffer) (audio.Buffer, error) {
	return f(ctx, b)
}

// DecodeFunc turns a stream of encoded audio into samples. The default is audio.Decode, which recognises every
// registered format by its leading bytes.
type DecodeFunc func(r io.Reader) (audio.SampleReader, error)

// ErrorHandler receives errors that do not stop a component, such as a failed Export. It must be safe for concurrent
// use and must not block.
type ErrorHandler func(err error)
