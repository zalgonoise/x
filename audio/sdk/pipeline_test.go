package sdk_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/internal/testutil"
	"github.com/zalgonoise/x/audio/sdk"
	"github.com/zalgonoise/x/audio/sdk/exporters"
	"github.com/zalgonoise/x/audio/sdk/sources"
)

const stereo16 = "amen_kick/amen_kick_stereo_16bit_44100hz.wav"

// decodeAll decodes a WAV file with the codec under its own tests, as the reference for what a pipeline must deliver.
func decodeAll(t *testing.T, b []byte) audio.Buffer {
	t.Helper()

	d, err := wav.NewDecoder(bytes.NewReader(b))
	require.NoError(t, err)

	buf, err := audio.ReadAll(d)
	require.NoError(t, err)

	return buf
}

func bytesSource(b []byte) sdk.Source {
	return sdk.SourceFunc(func(context.Context) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(b)), nil })
}

func TestPipelineFileToWAV(t *testing.T) {
	defer goleak.VerifyNone(t)

	file := fixtures.WAV(t, stereo16)
	want := decodeAll(t, file)

	in := filepath.Join(t.TempDir(), "in.wav")
	require.NoError(t, os.WriteFile(in, file, 0o600))

	out, err := os.Create(filepath.Join(t.TempDir(), "out.wav"))
	require.NoError(t, err)

	defer out.Close()

	rec := &recorder{}
	queued, err := sdk.NewQueuedExporter(exporters.NewWAV(out), sdk.WithPolicy(sdk.Block), sdk.WithQueueSize(2))
	require.NoError(t, err)

	p, err := sdk.NewPipeline(sources.NewFile(in), []sdk.Exporter{rec, queued},
		sdk.WithBufferDuration(10*time.Millisecond))
	require.NoError(t, err)

	require.NoError(t, p.Run(t.Context()), "a finite source ends Run with nil")
	require.NoError(t, p.Shutdown(t.Context()))

	// the recorder saw every frame once, in order, in 10 ms buffers
	var frame int64

	for i, b := range rec.buffers() {
		require.Equal(t, want.Format, b.Format)
		require.Equal(t, frame, b.Frame, "buffer %d", i)

		if i < len(rec.buffers())-1 {
			require.Equal(t, 441, b.Frames(), "10 ms at 44.1 kHz")
		}

		frame += int64(b.Frames())
	}

	require.Equal(t, want.Samples, rec.samples())

	// the queued WAV exporter wrote the same audio, with back-patched sizes
	written, err := os.ReadFile(out.Name())
	require.NoError(t, err)

	got := decodeAll(t, written)
	require.Equal(t, want.Format, got.Format)
	require.Equal(t, want.Samples, got.Samples)
	require.Zero(t, queued.Dropped())

	_, _, shutdowns := rec.counts()
	require.Equal(t, 1, shutdowns)
}

func TestPipelineTransformsAndErrors(t *testing.T) {
	file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16}.EncodeInts([][]int32{{1 << 14, -1 << 14, 1 << 13}})

	t.Run("TransformApplied", func(t *testing.T) {
		rec := &recorder{}
		gain := sdk.TransformFunc(func(_ context.Context, b audio.Buffer) (audio.Buffer, error) {
			out := b.Clone()
			for i := range out.Samples {
				out.Samples[i] *= 2
			}

			return out, nil
		})

		p, err := sdk.NewPipeline(bytesSource(file), []sdk.Exporter{rec}, sdk.WithTransforms(gain))
		require.NoError(t, err)
		require.NoError(t, p.Run(t.Context()))
		require.Equal(t, []float64{1, -1, 0.5}, rec.samples())
	})

	t.Run("TransformErrorStopsRun", func(t *testing.T) {
		boom := errors.New("boom")
		fail := sdk.TransformFunc(func(context.Context, audio.Buffer) (audio.Buffer, error) { return audio.Buffer{}, boom })

		p, err := sdk.NewPipeline(bytesSource(file), []sdk.Exporter{&recorder{}}, sdk.WithTransforms(fail))
		require.NoError(t, err)
		require.ErrorIs(t, p.Run(t.Context()), boom)
	})

	t.Run("ExportErrorsGoToTheHandler", func(t *testing.T) {
		boom := errors.New("boom")

		var (
			mu   sync.Mutex
			errs []error
		)

		failing := &recorder{exportErr: boom}
		healthy := &recorder{}

		p, err := sdk.NewPipeline(bytesSource(file), []sdk.Exporter{failing, healthy},
			sdk.WithBufferDuration(time.Nanosecond), // one frame per buffer
			sdk.WithErrorHandler(func(err error) {
				mu.Lock()
				defer mu.Unlock()

				errs = append(errs, err)
			}))
		require.NoError(t, err)
		require.NoError(t, p.Run(t.Context()), "export errors do not stop the pipeline")
		require.Len(t, healthy.samples(), 3, "later exporters still receive every buffer")
		require.Len(t, errs, 3)
		require.ErrorIs(t, errs[0], boom)
	})

	t.Run("SourceAndDecodeErrors", func(t *testing.T) {
		missing, err := sdk.NewPipeline(sources.NewFile(filepath.Join(t.TempDir(), "missing.wav")), nil)
		require.NoError(t, err)
		require.ErrorIs(t, missing.Run(t.Context()), os.ErrNotExist)

		notAudio, err := sdk.NewPipeline(bytesSource([]byte("definitely not a wav file")), nil)
		require.NoError(t, err)
		require.ErrorIs(t, notAudio.Run(t.Context()), audio.ErrUnknownFormat)

		custom, err := sdk.NewPipeline(bytesSource(file), nil, sdk.WithDecoder(func(r io.Reader) (audio.SampleReader, error) {
			return wav.NewDecoder(r)
		}))
		require.NoError(t, err)
		require.NoError(t, custom.Run(t.Context()))
	})
}

func TestNewPipelineValidation(t *testing.T) {
	src := bytesSource(nil)

	_, err := sdk.NewPipeline(nil, nil)
	require.ErrorIs(t, err, sdk.ErrNilSource)

	_, err = sdk.NewPipeline(src, []sdk.Exporter{&recorder{}, nil})
	require.ErrorIs(t, err, sdk.ErrNilExporter)

	for name, opt := range map[string]sdk.Option{
		"Logger":         sdk.WithLogger(nil),
		"ErrorHandler":   sdk.WithErrorHandler(nil),
		"Decoder":        sdk.WithDecoder(nil),
		"Transform":      sdk.WithTransforms(nil),
		"BufferDuration": sdk.WithBufferDuration(0),
	} {
		_, err = sdk.NewPipeline(src, nil, opt)
		require.ErrorIs(t, err, sdk.ErrInvalidOption, name)
	}
}

// blockingSource returns a stream that never delivers data after header until it is closed, like a stalled live
// stream. Every Open sends on the returned channel.
func blockingSource(header []byte) (sdk.Source, chan struct{}) {
	opened := make(chan struct{}, 16)

	return sdk.SourceFunc(func(context.Context) (io.ReadCloser, error) {
		pr, pw := io.Pipe()

		go pw.Write(header) //nolint:errcheck // the header, then nothing; fails once the pipeline closes the stream

		opened <- struct{}{}

		return pr, nil
	}), opened
}

func TestPipelineLifecycle(t *testing.T) {
	defer goleak.VerifyNone(t)

	header := testutil.LiveHeader(testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16}.Encode(nil), math.MaxUint32)

	t.Run("ShutdownStopsABlockedRun", func(t *testing.T) {
		src, opened := blockingSource(header)
		rec := &recorder{}

		p, err := sdk.NewPipeline(src, []sdk.Exporter{rec})
		require.NoError(t, err)

		ran := make(chan error, 1)

		go func() { ran <- p.Run(context.Background()) }()

		<-opened
		require.ErrorIs(t, p.Run(t.Context()), sdk.ErrRunning, "a second Run while the first is blocked")

		require.NoError(t, p.Shutdown(t.Context()))
		require.ErrorIs(t, <-ran, sdk.ErrShutdown, "the blocked read is unblocked by closing the stream")
		require.NoError(t, p.Shutdown(t.Context()), "idempotent")

		_, _, shutdowns := rec.counts()
		require.Equal(t, 1, shutdowns)
		require.ErrorIs(t, p.Run(t.Context()), sdk.ErrShutdown)
		require.ErrorIs(t, p.ForceFlush(t.Context()), sdk.ErrShutdown)
	})

	t.Run("CancelledContextAndRerun", func(t *testing.T) {
		src, opened := blockingSource(header)

		p, err := sdk.NewPipeline(src, nil)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
		defer cancel()

		require.ErrorIs(t, p.Run(ctx), context.DeadlineExceeded)

		ctx2, cancel2 := context.WithCancel(t.Context())
		cancel2()
		require.ErrorIs(t, p.Run(ctx2), context.Canceled, "a pipeline runs again after Run returned")
		require.Len(t, opened, 2, "each Run opens the source")
		require.NoError(t, p.Shutdown(t.Context()))
	})

	t.Run("ConcurrentShutdowns", func(t *testing.T) {
		src, _ := blockingSource(header)
		rec := &recorder{}

		p, err := sdk.NewPipeline(src, []sdk.Exporter{rec})
		require.NoError(t, err)

		ran := make(chan error, 1)

		go func() { ran <- p.Run(context.Background()) }()

		errs := make(chan error, 8)

		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() { errs <- p.Shutdown(context.Background()) })
		}

		wg.Wait()
		close(errs)

		for err := range errs {
			require.NoError(t, err)
		}

		err = <-ran
		require.True(t, err == nil || errors.Is(err, sdk.ErrShutdown), "Run returned %v", err)

		_, _, shutdowns := rec.counts()
		require.Equal(t, 1, shutdowns)
	})
}

// A live HTTP stream: the pipeline exports until Shutdown, which ends Run with ErrShutdown.
func TestPipelineHTTPStream(t *testing.T) {
	file := fixtures.WAV(t, stereo16)
	srv := testutil.NewWAVServer(t, file, testutil.Endless(math.MaxUint32), testutil.Chunked(1447))

	transport := &http.Transport{}
	defer transport.CloseIdleConnections()

	src, err := sources.NewHTTP(srv.URL, sources.WithClient(&http.Client{Transport: transport}))
	require.NoError(t, err)

	rec := &recorder{}
	queued, err := sdk.NewQueuedExporter(rec) // real time: drops the oldest buffers if the recorder falls behind
	require.NoError(t, err)

	p, err := sdk.NewPipeline(src, []sdk.Exporter{queued})
	require.NoError(t, err)

	ran := make(chan error, 1)

	go func() { ran <- p.Run(context.Background()) }()

	require.Eventually(t, func() bool {
		n, _, _ := rec.counts()

		return n >= 20
	}, 5*time.Second, time.Millisecond)

	require.NoError(t, p.Shutdown(t.Context()))
	require.ErrorIs(t, <-ran, sdk.ErrShutdown)

	want := decodeAll(t, file)
	got := rec.buffers()
	require.Equal(t, want.Format, got[0].Format)
	require.Equal(t, want.Samples[:len(got[0].Samples)], got[0].Samples, "the stream starts with the file's audio")
}
