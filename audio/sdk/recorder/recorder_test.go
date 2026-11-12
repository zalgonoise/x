package recorder_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/sdk"
	"github.com/zalgonoise/x/audio/sdk/recorder"
)

var mono = audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}

// byFrame names segments by their first frame only, so tests can predict the names.
var byFrame = recorder.WithNames(func(s recorder.Segment) string {
	return fmt.Sprintf("segment-%012d.wav", s.FirstFrame)
})

// memStore keeps objects in memory; failNext makes the next Put fail.
type memStore struct {
	mu       sync.Mutex
	objects  map[string][]byte
	order    []string
	failNext error
}

func (s *memStore) Put(_ context.Context, name string, r io.Reader, size int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.failNext; err != nil {
		s.failNext = nil

		return err
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if int64(len(b)) != size {
		return errors.New("size mismatch")
	}

	if s.objects == nil {
		s.objects = map[string][]byte{}
	}

	s.objects[name] = b
	s.order = append(s.order, name)

	return nil
}

// ramp returns a mono Buffer of n frames starting at frame, whose samples encode their frame index.
func ramp(frame int64, n int) audio.Buffer {
	s := make([]float64, n)
	for i := range s {
		s[i] = float64(frame+int64(i)) / 32768
	}

	return audio.Buffer{Format: mono, Samples: s, Frame: frame}
}

func decode(t *testing.T, b []byte) audio.Buffer {
	t.Helper()

	d, err := wav.NewDecoder(bytes.NewReader(b))
	require.NoError(t, err)

	buf, err := audio.ReadAll(d)
	require.NoError(t, err)

	return buf
}

func TestRecorderSegments(t *testing.T) {
	store := &memStore{}

	r, err := recorder.New(store, byFrame, recorder.WithSegmentDuration(100*time.Millisecond)) // 800 frames
	require.NoError(t, err)

	for i := range 9 { // 2700 frames in buffers of 300, which straddle segment boundaries
		require.NoError(t, r.Export(t.Context(), ramp(int64(i*300), 300)))
	}

	require.Len(t, store.order, 3, "full segments are stored as soon as they fill up")
	require.NoError(t, r.Shutdown(t.Context()))
	require.NoError(t, r.Shutdown(t.Context()), "idempotent")

	require.Equal(t, []string{
		"segment-000000000000.wav", "segment-000000000800.wav", "segment-000000001600.wav", "segment-000000002400.wav",
	}, store.order)

	var all []float64

	for i, name := range store.order {
		buf := decode(t, store.objects[name])
		require.Equal(t, mono, buf.Format)

		if i < 3 {
			require.Equal(t, 800, buf.Frames(), name)
		}

		all = append(all, buf.Samples...)
	}

	require.Equal(t, ramp(0, 2700).Samples, all, "the segments hold the whole stream, in order")
	require.ErrorIs(t, r.Export(t.Context(), ramp(2700, 1)), sdk.ErrShutdown)
	require.ErrorIs(t, r.ForceFlush(t.Context()), sdk.ErrShutdown)
}

func TestRecorderFlushAndFormatChange(t *testing.T) {
	store := &memStore{}

	r, err := recorder.New(store, byFrame)
	require.NoError(t, err)

	require.NoError(t, r.ForceFlush(t.Context()), "nothing to flush")
	require.NoError(t, r.Export(t.Context(), ramp(0, 10)))
	require.NoError(t, r.ForceFlush(t.Context()))
	require.Equal(t, []string{"segment-000000000000.wav"}, store.order, "the partial segment is stored")

	require.NoError(t, r.Export(t.Context(), ramp(10, 10)))

	stereo := audio.Buffer{
		Format:  audio.Format{SampleRate: 8000, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM},
		Samples: []float64{0.5, -0.5},
		Frame:   20,
	}
	require.NoError(t, r.Export(t.Context(), stereo), "a format change starts a new segment")
	require.NoError(t, r.Export(t.Context(), audio.Buffer{Format: mono}), "an empty buffer is ignored")
	require.NoError(t, r.Shutdown(t.Context()))

	require.Equal(t, []string{
		"segment-000000000000.wav", "segment-000000000010.wav", "segment-000000000020.wav",
	}, store.order)
	require.Equal(t, ramp(10, 10).Samples, decode(t, store.objects["segment-000000000010.wav"]).Samples)
	require.Equal(t, 2, decode(t, store.objects["segment-000000000020.wav"]).Format.Channels)
}

func TestRecorderErrors(t *testing.T) {
	boom := errors.New("boom")

	t.Run("StoreError", func(t *testing.T) {
		store := &memStore{failNext: boom}

		r, err := recorder.New(store, byFrame, recorder.WithSegmentDuration(time.Millisecond)) // 8 frames
		require.NoError(t, err)

		require.ErrorIs(t, r.Export(t.Context(), ramp(0, 8)), boom, "the failed segment is reported")
		require.NoError(t, r.Export(t.Context(), ramp(8, 8)), "recording goes on")
		require.Equal(t, []string{"segment-000000000008.wav"}, store.order)
	})

	t.Run("EncoderError", func(t *testing.T) {
		r, err := recorder.New(&memStore{}, recorder.WithEncoder(func(io.Writer, audio.Format) (recorder.Encoder, error) {
			return nil, boom
		}))
		require.NoError(t, err)
		require.ErrorIs(t, r.Export(t.Context(), ramp(0, 8)), boom)
	})

	t.Run("UnsupportedFormat", func(t *testing.T) {
		r, err := recorder.New(&memStore{})
		require.NoError(t, err)

		bad := audio.Buffer{Format: audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16}, Samples: []float64{0}}
		require.ErrorIs(t, r.Export(t.Context(), bad), audio.ErrInvalidFormat)
	})

	t.Run("Options", func(t *testing.T) {
		_, err := recorder.New(nil)
		require.ErrorIs(t, err, sdk.ErrInvalidOption)

		for name, opt := range map[string]recorder.Option{
			"SegmentDuration": recorder.WithSegmentDuration(0),
			"Encoder":         recorder.WithEncoder(nil),
			"Names":           recorder.WithNames(nil),
			"Clock":           recorder.WithClock(nil),
		} {
			_, err := recorder.New(&memStore{}, opt)
			require.ErrorIs(t, err, sdk.ErrInvalidOption, name)
		}
	})

	t.Run("CustomNames", func(t *testing.T) {
		store := &memStore{}

		r, err := recorder.New(store, recorder.WithNames(func(s recorder.Segment) string {
			return "take/" + s.Format.Encoding.String() + "-" + string(rune('a'+s.FirstFrame))
		}))
		require.NoError(t, err)
		require.NoError(t, r.Export(t.Context(), ramp(1, 4)))
		require.NoError(t, r.Shutdown(t.Context()))
		require.Equal(t, []string{"take/pcm-b"}, store.order)
	})
}

func TestDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rec")

	d, err := recorder.NewDir(path)
	require.NoError(t, err)

	defer d.Close()

	require.NoError(t, d.Put(t.Context(), "a.wav", bytes.NewReader([]byte("one")), 3))
	require.NoError(t, d.Put(t.Context(), "a.wav", bytes.NewReader([]byte("two")), 3), "objects are replaced")
	require.NoError(t, d.Put(t.Context(), "day/b.wav", bytes.NewReader([]byte("x")), 1), "subdirectories are created")

	b, err := os.ReadFile(filepath.Join(path, "a.wav"))
	require.NoError(t, err)
	require.Equal(t, "two", string(b))

	b, err = os.ReadFile(filepath.Join(path, "day", "b.wav"))
	require.NoError(t, err)
	require.Equal(t, "x", string(b))

	for _, bad := range []string{"../escape.wav", "/abs.wav", "", "a/../../b"} {
		require.ErrorIs(t, d.Put(t.Context(), bad, bytes.NewReader(nil), 0), recorder.ErrInvalidName, bad)
	}

	require.ErrorIs(t, d.Put(t.Context(), "short.wav", bytes.NewReader([]byte("ab")), 3), recorder.ErrShortObject)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorIs(t, d.Put(ctx, "c.wav", bytes.NewReader(nil), 0), context.Canceled)

	entries, err := os.ReadDir(path)
	require.NoError(t, err)

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}

	slices.Sort(names)
	require.Equal(t, []string{"a.wav", "day"}, names, "no partial or temporary files are left behind")
}

// The recorder in a pipeline, behind a blocking queue: the stored segments add up to the source.
func TestRecorderInPipeline(t *testing.T) {
	file := fixtures.WAV(t, "amen_kick/amen_kick_stereo_16bit_44100hz.wav")
	want := decode(t, file)

	path := t.TempDir()

	store, err := recorder.NewDir(path)
	require.NoError(t, err)

	defer store.Close()

	rec, err := recorder.New(store, recorder.WithSegmentDuration(50*time.Millisecond))
	require.NoError(t, err)

	queued, err := sdk.NewQueuedExporter(rec, sdk.WithPolicy(sdk.Block))
	require.NoError(t, err)

	p, err := sdk.NewPipeline(sdk.SourceFunc(func(context.Context) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(file)), nil
	}), []sdk.Exporter{queued})
	require.NoError(t, err)
	require.NoError(t, p.Run(t.Context()))
	require.NoError(t, p.Shutdown(t.Context()))

	entries, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, entries, (want.Frames()+2204)/2205, "50 ms segments at 44.1 kHz")

	all := make([]float64, 0, len(want.Samples))

	for _, e := range entries { // ReadDir sorts by name, which is stream order
		b, err := os.ReadFile(filepath.Join(path, e.Name()))
		require.NoError(t, err)

		all = append(all, decode(t, b).Samples...)
	}

	require.Equal(t, want.Samples, all)
}

// Review round 3 (2026-10-27): a segment only holds contiguous frames, and default names never collide across
// reconnects.
func TestRecorderContiguousSegments(t *testing.T) {
	store := &memStore{}

	r, err := recorder.New(store, byFrame, recorder.WithSegmentDuration(time.Second))
	require.NoError(t, err)

	require.NoError(t, r.Export(t.Context(), ramp(0, 50)))
	require.NoError(t, r.Export(t.Context(), ramp(1000, 50)), "a gap (frames dropped upstream)")
	require.NoError(t, r.Export(t.Context(), ramp(0, 20)), "a restart from frame 0 (a reconnect)")
	require.NoError(t, r.Shutdown(t.Context()))

	require.Equal(t, []string{"segment-000000000000.wav", "segment-000000001000.wav", "segment-000000000000.wav"},
		store.order, "each run of contiguous frames is its own segment")
	require.Equal(t, ramp(0, 20).Samples, decode(t, store.objects["segment-000000000000.wav"]).Samples)
}

func TestRecorderDefaultNames(t *testing.T) {
	store := &memStore{}
	now := time.Date(2026, 10, 27, 3, 15, 0, 250_000_000, time.FixedZone("CET", 3600))

	r, err := recorder.New(store, recorder.WithSegmentDuration(time.Millisecond), recorder.WithClock(func() time.Time {
		now = now.Add(time.Millisecond)

		return now
	}))
	require.NoError(t, err)

	for range 2 { // two "runs" restarting at frame 0
		require.NoError(t, r.Export(t.Context(), ramp(0, 8)))
	}

	require.NoError(t, r.Shutdown(t.Context()))
	require.Equal(t, []string{"20261027T021500.251Z_000000000000.wav", "20261027T021500.252Z_000000000000.wav"},
		store.order, "UTC start time first: names sort by time and do not collide")
}

func TestDirRejectsNonFileNames(t *testing.T) {
	path := t.TempDir()

	d, err := recorder.NewDir(path)
	require.NoError(t, err)

	defer d.Close()

	for _, bad := range []string{".", "d/.", "a/", "x/../b.wav", "./c.wav", "a//b.wav"} {
		require.ErrorIs(t, d.Put(t.Context(), bad, bytes.NewReader(nil), 0), recorder.ErrInvalidName, bad)
	}

	entries, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Empty(t, entries, "rejected names leave nothing behind")
}
