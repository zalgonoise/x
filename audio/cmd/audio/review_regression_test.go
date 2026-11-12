package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"math"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/internal/testutil"
)

// Review of the spectrum code (2026-11-02).

// jsonLines decodes the JSON lines on stdout, failing on any line that is not a JSON object.
func jsonLines(t *testing.T, out string) []map[string]any {
	t.Helper()

	var recs []map[string]any

	sc := bufio.NewScanner(strings.NewReader(out))
	sc.Buffer(nil, 1<<24)

	for sc.Scan() {
		var rec map[string]any

		require.NoError(t, json.Unmarshal(sc.Bytes(), &rec), sc.Text())

		recs = append(recs, rec)
	}

	require.NoError(t, sc.Err())

	return recs
}

// A file decodes far faster than real time: the meter's queue blocks instead of dropping buffers, so every one is
// measured. Every sample of this input is clipped, so audio.clipped_samples adds up to the frames of each channel.
func TestFileInputMeasuresEveryBuffer(t *testing.T) {
	const rate, secs = 44100, 10

	n := rate * secs
	square := make([]int32, n)

	for i := range square {
		square[i] = math.MaxInt16
		if (i/50)%2 == 1 {
			square[i] = math.MinInt16
		}
	}

	file := testutil.WAV{Channels: 2, SampleRate: rate, BitsPerSample: 16}.EncodeInts([][]int32{square, square})

	var stdout, stderr syncBuffer

	code := run(t.Context(), []string{"-input", "-", "-queue", "1"}, env(nil), bytes.NewReader(file), &stdout, &stderr)
	require.Equal(t, exitOK, code, stderr.String())
	require.NotContains(t, stderr.String(), "fell behind")

	clipped := map[float64]float64{}

	for _, rec := range jsonLines(t, stdout.String()) {
		if rec["msg"] != "audio.clipped_samples" {
			continue
		}

		ch, ok := rec["channel"].(float64)
		require.True(t, ok)

		v, ok := rec["value"].(float64)
		require.True(t, ok)

		clipped[ch] += v
	}

	require.Equal(t, map[float64]float64{0: float64(n), 1: float64(n)}, clipped)
}

// A header declaring many channels costs little memory until each channel carries a whole frame.
func TestManyChannelsBoundedMemory(t *testing.T) {
	const channels = 1024

	perChannel := make([][]int32, channels)
	for i := range perChannel {
		perChannel[i] = []int32{0} // one frame of silence
	}

	file := testutil.WAV{Channels: channels, SampleRate: 44100, BitsPerSample: 8}.EncodeInts(perChannel)

	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	var stdout, stderr syncBuffer

	code := run(t.Context(), []string{"-input", "-"}, env(nil), bytes.NewReader(file), &stdout, &stderr)

	runtime.ReadMemStats(&after)

	require.Equal(t, exitOK, code, stderr.String())
	require.Less(t, after.TotalAlloc-before.TotalAlloc, uint64(64<<20), "a %d-byte input", len(file))
}

// Lines keep their shape whatever the audio: a NaN sample in a float WAV is logged as the string "NaN".
func TestNonFiniteValuesKeepTheLineShape(t *testing.T) {
	samples := testutil.Sine(8192, 44100, 1000, 0.5, 0)
	samples[100] = math.NaN()

	file := testutil.WAV{Format: testutil.FormatFloat, Channels: 1, SampleRate: 44100, BitsPerSample: 32}.
		EncodeFloats([][]float64{samples})

	var stdout, stderr syncBuffer

	code := run(t.Context(), []string{"-input", "-"}, env(nil), bytes.NewReader(file), &stdout, &stderr)
	require.Equal(t, exitOK, code, stderr.String())

	var spectra, peaks int

	for _, rec := range jsonLines(t, stdout.String()) {
		msg, ok := rec["msg"].(string)
		require.True(t, ok)

		switch {
		case strings.HasPrefix(msg, "audio.spectrum."):
			spectra++

			values, ok := rec["values"].([]any)
			require.True(t, ok, "%s: values = %#v", msg, rec["values"])
			require.Contains(t, values, "NaN", "the frames holding the NaN sample")
		case msg == "audio.peak":
			peaks++

			require.Equal(t, "NaN", rec["value"])
		}
	}

	require.Positive(t, spectra)
	require.Positive(t, peaks)
}
