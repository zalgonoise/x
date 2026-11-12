package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

const stereo16 = "amen_kick/amen_kick_stereo_16bit_44100hz.wav"

// syncBuffer is a bytes.Buffer safe for the concurrent writes of loggers and readers.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.b.String()
}

func decode(t *testing.T, b []byte) audio.Buffer {
	t.Helper()

	d, err := wav.NewDecoder(bytes.NewReader(b))
	require.NoError(t, err)

	buf, err := audio.ReadAll(d)
	require.NoError(t, err)

	return buf
}

// From stdin: per-channel levels on stdout, the stream recorded in segments, exit 0 when the input ends.
func TestRunStdinMetricsAndRecording(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	file := fixtures.WAV(t, stereo16)
	dir := t.TempDir()

	var stdout, stderr syncBuffer

	code := run(t.Context(), []string{"-input", "-", "-record.dir", dir, "-record.segment", "200ms"}, env(nil),
		bytes.NewReader(file), &stdout, &stderr)
	require.Equal(t, exitOK, code, stderr.String())
	require.Contains(t, stderr.String(), "input ended")

	// the final batch is emitted on shutdown: every level, for both channels
	seen := map[string]map[float64]bool{}

	sc := bufio.NewScanner(strings.NewReader(stdout.String()))
	for sc.Scan() {
		var rec struct {
			Msg     string  `json:"msg"`
			Channel float64 `json:"channel"`
		}

		require.NoError(t, json.Unmarshal(sc.Bytes(), &rec), sc.Text())

		if seen[rec.Msg] == nil {
			seen[rec.Msg] = map[float64]bool{}
		}

		seen[rec.Msg][rec.Channel] = true
	}

	for _, name := range []string{
		"audio.peak", "audio.rms", "audio.crest_factor", "audio.dc_offset",
		"audio.zero_crossing_rate", "audio.clipped_samples", "audio.spectrum.peak", "audio.spectrum.mean",
	} {
		require.Equal(t, map[float64]bool{0: true, 1: true}, seen[name], name)
	}

	// the segments add up to the input
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	all := make([]float64, 0, len(file)/2)

	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		require.NoError(t, err)

		all = append(all, decode(t, b).Samples...)
	}

	require.Equal(t, decode(t, file).Samples, all)
}

// With -metrics prometheus the levels are served on /metrics while the stream plays.
func TestRunPrometheus(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	transport := &http.Transport{}
	defer transport.CloseIdleConnections()

	client := &http.Client{Transport: transport}

	// a live stream: its data runs until the input closes, so the endpoint stays up while it is scraped
	file := testutil.LiveHeader(fixtures.WAV(t, stereo16), math.MaxUint32)
	pr, pw := io.Pipe()

	var stdout, stderr syncBuffer

	done := make(chan int, 1)

	go func() {
		done <- run(t.Context(), []string{"-input", "-", "-metrics", "prometheus", "-prometheus.addr", "127.0.0.1:0"},
			env(nil), pr, &stdout, &stderr)
	}()

	// stream the file, keeping the input open until the endpoint was scraped
	go func() { _, _ = pw.Write(file) }() //nolint:errcheck // a failed write shows up as a failed scrape

	addr := regexp.MustCompile(`addr=(\S+)`)

	var url string

	require.Eventually(t, func() bool {
		m := addr.FindStringSubmatch(stderr.String())
		if m == nil {
			return false
		}

		url = "http://" + m[1] + "/metrics"

		return true
	}, 5*time.Second, 5*time.Millisecond)

	require.Eventually(t, func() bool {
		res, err := client.Get(url)
		if err != nil {
			return false
		}

		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)

		return err == nil && strings.Contains(string(b), `audio_peak{channel="1"}`) &&
			strings.Contains(string(b), `audio_spectrum_peak{channel="1",frequency="1000"}`) &&
			strings.Contains(string(b), `audio_spectrum_mean{channel="0",frequency="31.5"}`) &&
			strings.Contains(string(b), "go_goroutines")
	}, 5*time.Second, 10*time.Millisecond)

	require.NoError(t, pw.Close())
	require.Equal(t, exitOK, <-done, stderr.String())
	require.Empty(t, stdout.String(), "prometheus mode writes nothing on stdout")
}

// -metrics.spectrum picks the layout of the logged spectra: bands, every FFT bin, or none.
func TestRunSpectrumLayouts(t *testing.T) {
	file := fixtures.WAV(t, stereo16)

	for _, tc := range []struct {
		args []string
		bins int // frequencies in each spectrum line; 0 when no spectrum is logged
	}{
		{args: []string{"-metrics.spectrum", "octave"}, bins: 10},       // 31.5 Hz to 16 kHz
		{args: []string{"-metrics.spectrum", "third-octave"}, bins: 30}, // 20 Hz to 20 kHz but 25 Hz: no bin at 10.8 Hz
		{args: []string{"-metrics.spectrum", "full", "-metrics.spectrum.size", "256"}, bins: 129},
		{args: []string{"-metrics.spectrum", "none"}},
	} {
		t.Run(tc.args[1], func(t *testing.T) {
			var stdout, stderr syncBuffer

			code := run(t.Context(), append([]string{"-input", "-"}, tc.args...), env(nil), bytes.NewReader(file),
				&stdout, &stderr)
			require.Equal(t, exitOK, code, stderr.String())

			var spectra int

			sc := bufio.NewScanner(strings.NewReader(stdout.String()))
			sc.Buffer(nil, 1<<20)

			for sc.Scan() {
				var rec struct {
					Msg         string    `json:"msg"`
					Frequencies []float64 `json:"frequencies"`
					Values      []float64 `json:"values"`
				}

				require.NoError(t, json.Unmarshal(sc.Bytes(), &rec), sc.Text())

				if strings.HasPrefix(rec.Msg, "audio.spectrum.") {
					spectra++

					require.Len(t, rec.Frequencies, tc.bins, rec.Msg)
					require.Len(t, rec.Values, tc.bins, rec.Msg)
				}
			}

			require.NoError(t, sc.Err())

			if tc.bins == 0 {
				require.Zero(t, spectra)
			} else {
				require.Positive(t, spectra)
			}
		})
	}
}

func TestRunErrors(t *testing.T) {
	var stderr syncBuffer

	require.Equal(t, exitUsage, run(t.Context(), nil, env(nil), nil, io.Discard, &stderr), "no input")
	require.Contains(t, stderr.String(), "-input is required")

	require.Equal(t, exitOK, run(t.Context(), []string{"-h"}, env(nil), nil, io.Discard, io.Discard))

	require.Equal(t, exitFailure, run(t.Context(), []string{"-input", "-", "-metrics", "none"}, env(nil),
		strings.NewReader("not a wav file"), io.Discard, io.Discard), "undecodable input")

	busy := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(busy, nil, 0o600))
	require.Equal(t, exitFailure, run(t.Context(), []string{"-input", "-", "-record.dir", filepath.Join(busy, "sub")},
		env(nil), bytes.NewReader(nil), io.Discard, io.Discard), "the record directory cannot be created")

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	pr, pw := io.Pipe()
	defer pw.Close()

	require.Equal(t, exitOK, run(ctx, []string{"-input", "-", "-metrics", "none"}, env(nil), pr, io.Discard, io.Discard),
		"a signal (cancelled context) is a normal exit")
}

// -duration bounds a live stream that never ends.
func TestRunDuration(t *testing.T) {
	pr, pw := io.Pipe()
	defer pw.Close()

	go func() { _, _ = pw.Write(fixtures.WAV(t, stereo16)[:44]) }() //nolint:errcheck // the header, then silence

	var stderr syncBuffer

	start := time.Now()
	code := run(t.Context(), []string{"-input", "-", "-duration", "200ms", "-metrics", "none"}, env(nil), pr,
		io.Discard, &stderr)
	require.Equal(t, exitOK, code, stderr.String())
	require.Less(t, time.Since(start), 5*time.Second)
	require.Contains(t, stderr.String(), "run duration reached")
}

// -reconnect runs a live HTTP stream again after it drops, until -duration.
func TestRunReconnect(t *testing.T) {
	file := fixtures.WAV(t, stereo16)
	srv := testutil.NewWAVServer(t, file, testutil.Endless(math.MaxUint32), testutil.DisconnectAfter(20000))

	var stderr syncBuffer

	code := run(t.Context(), []string{
		"-input", srv.URL, "-metrics", "none", "-duration", "500ms",
		"-reconnect", "-reconnect.min-wait", "10ms", "-reconnect.max-wait", "20ms",
	}, env(nil), nil, io.Discard, &stderr)
	require.Equal(t, exitOK, code, stderr.String())
	require.Contains(t, stderr.String(), "reconnecting")
	require.GreaterOrEqual(t, srv.Requests(), int64(2))

	// without -reconnect, the dropped connection is a failure
	code = run(t.Context(), []string{"-input", srv.URL, "-metrics", "none", "-duration", "5s"}, env(nil), nil,
		io.Discard, io.Discard)
	require.Equal(t, exitFailure, code)
}
