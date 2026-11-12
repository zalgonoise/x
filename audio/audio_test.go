package audio_test

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio"
)

func TestFormat(t *testing.T) {
	f := audio.Format{SampleRate: 48000, Channels: 2, BitDepth: 24, Encoding: audio.EncodingPCM}

	require.NoError(t, f.Validate())
	require.Equal(t, 3, f.SampleSize())
	require.Equal(t, 6, f.BlockAlign())
	require.Equal(t, 288000, f.ByteRate())
	require.Equal(t, 24, f.Significant())
	require.Equal(t, time.Second, f.Duration(48000))
	require.Equal(t, 1500*time.Millisecond, f.Duration(72000))
	require.Equal(t, int64(24000), f.Frames(500*time.Millisecond))
	require.Equal(t, int64(4*3600*48000), f.Frames(4*time.Hour), "long durations do not overflow")
	require.Equal(t, "pcm 24-bit 2ch 48000Hz", f.String())

	f.BitDepth, f.ValidBits = 32, 20
	require.Equal(t, 20, f.Significant())
	require.Equal(t, "pcm 20-in-32-bit 2ch 48000Hz", f.String())

	for name, bad := range map[string]audio.Format{
		"NoRate":          {Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM},
		"NoChannels":      {SampleRate: 8000, BitDepth: 16, Encoding: audio.EncodingPCM},
		"PartialByte":     {SampleRate: 8000, Channels: 1, BitDepth: 12, Encoding: audio.EncodingPCM},
		"TooManyValid":    {SampleRate: 8000, Channels: 1, BitDepth: 16, ValidBits: 20, Encoding: audio.EncodingPCM},
		"UnknownEncoding": {SampleRate: 8000, Channels: 1, BitDepth: 16},
	} {
		require.ErrorIs(t, bad.Validate(), audio.ErrInvalidFormat, name)
	}
}

func TestBuffer(t *testing.T) {
	b := audio.Buffer{
		Format:  audio.Format{SampleRate: 4, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM},
		Samples: []float64{1, -1, 0.5, 0.25, 0, 0},
		Frame:   8,
	}

	require.Equal(t, 3, b.Frames())
	require.Equal(t, 750*time.Millisecond, b.Duration())
	require.Equal(t, 2*time.Second, b.Offset())
	require.Equal(t, []float64{1, 0.5, 0}, b.Channel(0, nil))
	require.Equal(t, []float64{-1, 0.25, 0}, b.Channel(1, nil))
	require.Nil(t, b.Channel(2, nil))
	require.Equal(t, []float64{0, 0.375, 0}, b.Mixdown(nil))

	c := b.Clone()
	c.Samples[0] = 42
	require.InDelta(t, 1.0, b.Samples[0], 0, "Clone does not share samples")
}

// sliceReader serves samples in small reads and returns the last ones together with io.EOF.
type sliceReader struct {
	f       audio.Format
	samples []float64
}

func (r *sliceReader) Format() audio.Format { return r.f }

func (r *sliceReader) ReadSamples(dst []float64) (int, error) {
	n := copy(dst[:min(len(dst), 2)], r.samples)
	r.samples = r.samples[n:]

	if len(r.samples) == 0 {
		return n, io.EOF
	}

	return n, nil
}

func TestReadAll(t *testing.T) {
	f := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}

	buf, err := audio.ReadAll(&sliceReader{f: f, samples: []float64{1, 2, 3, 4, 5}})
	require.NoError(t, err)
	require.Equal(t, f, buf.Format)
	require.Equal(t, []float64{1, 2, 3, 4, 5}, buf.Samples, "samples returned with io.EOF are kept")
}

// fakeReader is a SampleReader whose ReadSamples is a function.
type fakeReader struct {
	f    audio.Format
	read func(dst []float64) (int, error)
}

func (r fakeReader) Format() audio.Format                   { return r.f }
func (r fakeReader) ReadSamples(dst []float64) (int, error) { return r.read(dst) }

func TestReadAllBoundsItsBuffer(t *testing.T) {
	// the channel count comes from untrusted headers: 65535 channels must not mean 65535 × 4096 samples per read
	f := audio.Format{SampleRate: 8000, Channels: 65535, BitDepth: 8, Encoding: audio.EncodingPCM}

	var asked []int

	buf, err := audio.ReadAll(fakeReader{f: f, read: func(dst []float64) (int, error) {
		asked = append(asked, len(dst))
		if len(asked) == 1 {
			return f.Channels, nil
		}

		return 0, io.EOF
	}})
	require.NoError(t, err)
	require.Len(t, buf.Samples, f.Channels)
	require.Equal(t, []int{f.Channels, f.Channels}, asked, "one whole frame per read, the minimum")

	mono := audio.Format{SampleRate: 8000, Channels: 3, BitDepth: 8, Encoding: audio.EncodingPCM}
	_, err = audio.ReadAll(fakeReader{f: mono, read: func(dst []float64) (int, error) {
		require.Zero(t, len(dst)%3, "whole frames")
		require.LessOrEqual(t, len(dst), 1<<15)

		return 0, io.EOF
	}})
	require.NoError(t, err)
}

func TestReadAllNoProgress(t *testing.T) {
	f := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}
	calls := 0

	_, err := audio.ReadAll(fakeReader{f: f, read: func([]float64) (int, error) {
		calls++

		return 0, nil
	}})
	require.ErrorIs(t, err, io.ErrNoProgress)
	require.Equal(t, 100, calls)

	// empty reads interleaved with progress are fine
	calls = 0
	buf, err := audio.ReadAll(fakeReader{f: f, read: func(dst []float64) (int, error) {
		calls++

		switch {
		case calls > 300:
			return 0, io.EOF
		case calls%99 == 0:
			dst[0] = 1

			return 1, nil
		default:
			return 0, nil
		}
	}})
	require.NoError(t, err)
	require.Len(t, buf.Samples, 3)
}

func TestFormatDurationFramesRoundTrip(t *testing.T) {
	for _, rate := range []int{1, 8000, 44100, 48000, 96000, 192000, 352800, 1_000_000_000} {
		f := audio.Format{SampleRate: rate, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}

		for _, n := range []int64{0, 1, 2, 3, 7, 999, int64(rate) - 1, int64(rate), int64(rate) + 1, 3*int64(rate) + 5} {
			require.Equal(t, n, f.Frames(f.Duration(n)), "%d frames at %d Hz", n, rate)
		}
	}

	f := audio.Format{SampleRate: 44100, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}
	for n := range int64(44101) {
		if f.Frames(f.Duration(n)) != n {
			t.Fatalf("Frames(Duration(%d)) = %d at 44100 Hz", n, f.Frames(f.Duration(n)))
		}
	}

	require.Equal(t, 22676*time.Nanosecond, f.Duration(1), "rounded up from 22675.7 ns")
	require.Equal(t, time.Second, f.Duration(44100), "exact durations stay exact")
	require.Zero(t, f.Duration(-1))

	b := audio.Buffer{Format: f, Frame: 12345}
	require.Equal(t, b.Frame, f.Frames(b.Offset()), "Offset maps back to Frame")
}

func TestFormatValidateEncodingRange(t *testing.T) {
	for _, e := range []audio.Encoding{audio.EncodingUnknown, audio.EncodingMuLaw + 1, 42} {
		f := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: e}
		require.ErrorIs(t, f.Validate(), audio.ErrInvalidFormat, "%s", e)
		require.False(t, e.Valid())
	}

	for _, e := range []audio.Encoding{audio.EncodingPCM, audio.EncodingFloat, audio.EncodingALaw, audio.EncodingMuLaw} {
		require.True(t, e.Valid(), "%s", e)
	}
}

func TestBufferWithoutChannels(t *testing.T) {
	b := audio.Buffer{Samples: []float64{1, 2, 3}}
	require.Zero(t, b.Frames())
	require.Empty(t, b.Mixdown(nil), "one value per frame, and there are none")
	require.Empty(t, b.Channel(0, nil))
}

var errConnReset = errors.New("connection reset")

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errConnReset }

func TestDecodeErrors(t *testing.T) {
	_, _, err := audio.Decode(failingReader{})
	require.ErrorIs(t, err, errConnReset, "an I/O error is not an unknown format")
	require.NotErrorIs(t, err, audio.ErrUnknownFormat)

	_, _, err = audio.Decode(nil)
	require.ErrorIs(t, err, audio.ErrUnknownFormat)

	_, _, err = audio.Decode(bytes.NewReader(nil))
	require.ErrorIs(t, err, audio.ErrUnknownFormat, "empty input")
}

func TestRegistry(t *testing.T) {
	f := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 8, Encoding: audio.EncodingPCM}

	audio.RegisterFormat("zz-test", "ZZ??TEST", func(r io.Reader) (audio.SampleReader, error) {
		head := make([]byte, 8)
		if _, err := io.ReadFull(r, head); err != nil {
			return nil, err
		}

		return &sliceReader{f: f, samples: []float64{float64(head[2])}}, nil
	})

	sr, name, err := audio.Decode(bytes.NewReader([]byte("ZZ\x07\x00TEST...")))
	require.NoError(t, err)
	require.Equal(t, "zz-test", name)

	buf, err := audio.ReadAll(sr)
	require.NoError(t, err)
	require.Equal(t, []float64{7}, buf.Samples, "the decoder sees the stream from its first byte")

	_, _, err = audio.Decode(bytes.NewReader([]byte("nothing registered matches this")))
	require.True(t, errors.Is(err, audio.ErrUnknownFormat))

	_, _, err = audio.Decode(bytes.NewReader([]byte("ZZ")))
	require.ErrorIs(t, err, audio.ErrUnknownFormat, "input shorter than the magic")
}
