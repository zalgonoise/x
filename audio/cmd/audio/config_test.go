package main

import (
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func env(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

func TestParseConfig(t *testing.T) {
	file := filepath.Join(t.TempDir(), "in.wav")
	require.NoError(t, os.WriteFile(file, []byte("RIFF"), 0o600))

	t.Run("Defaults", func(t *testing.T) {
		c, err := parseConfig([]string{"-input", "-"}, env(nil), io.Discard)
		require.NoError(t, err)
		require.Equal(t, "-", c.input)
		require.Zero(t, c.duration, "runs until the input ends")
		require.Equal(t, 100*time.Millisecond, c.buffer)
		require.Equal(t, metricsLog, c.metrics)
		require.Equal(t, time.Second, c.interval)
		require.Equal(t, time.Minute, c.segment)
		require.Equal(t, 64, c.queue)
		require.Equal(t, slog.LevelInfo, c.logLevel)
		require.False(t, c.logJSON)
		require.Empty(t, c.recordDir)
		require.Equal(t, spectrumThirdOctave, c.spectrum)
		require.Equal(t, 4096, c.specSize)
	})

	t.Run("EnvFallbackAndFlagPrecedence", func(t *testing.T) {
		c, err := parseConfig([]string{"-metrics", "none"}, env(map[string]string{
			"AUDIO_INPUT":                 "file://" + file,
			"AUDIO_METRICS":               "prometheus",
			"AUDIO_METRICS_INTERVAL":      "250ms",
			"AUDIO_RECORD_DIR":            "/tmp/rec",
			"AUDIO_LOG_LEVEL":             "debug",
			"AUDIO_LOG_FORMAT":            "json",
			"AUDIO_METRICS_SPECTRUM":      "octave",
			"AUDIO_METRICS_SPECTRUM_SIZE": "8192",
		}), io.Discard)
		require.NoError(t, err)
		require.Equal(t, "file://"+file, c.input)
		require.Equal(t, metricsNone, c.metrics, "the flag wins over the environment")
		require.Equal(t, 250*time.Millisecond, c.interval)
		require.Equal(t, "/tmp/rec", c.recordDir)
		require.Equal(t, slog.LevelDebug, c.logLevel)
		require.True(t, c.logJSON)
		require.Equal(t, spectrumOctave, c.spectrum)
		require.Equal(t, 8192, c.specSize)
	})

	for name, tc := range map[string]struct {
		args []string
		env  map[string]string
		want error
	}{
		"MissingInput":     {want: ErrMissingInput},
		"UnexpectedArg":    {args: []string{"-input", "-", "extra"}, want: ErrUnexpectedArg},
		"BadEnvDuration":   {args: []string{"-input", "-"}, env: map[string]string{"AUDIO_DURATION": "soon"}, want: ErrInvalidEnv},
		"NegativeDuration": {args: []string{"-input", "-", "-duration", "-1s"}, want: ErrInvalidValue},
		"ZeroBuffer":       {args: []string{"-input", "-", "-buffer", "0s"}, want: ErrInvalidValue},
		"ZeroInterval":     {args: []string{"-input", "-", "-metrics.interval", "0s"}, want: ErrInvalidValue},
		"ZeroSegment":      {args: []string{"-input", "-", "-record.segment", "0s"}, want: ErrInvalidValue},
		"ZeroQueue":        {args: []string{"-input", "-", "-queue", "0"}, want: ErrInvalidValue},
		"ZeroMinWait":      {args: []string{"-input", "-", "-reconnect.min-wait", "0s"}, want: ErrInvalidValue},
		"MaxBelowMin":      {args: []string{"-input", "-", "-reconnect.max-wait", "1ms"}, want: ErrInvalidValue},
		"BadMetrics":       {args: []string{"-input", "-", "-metrics", "graphite"}, want: ErrInvalidValue},
		"BadSpectrum":      {args: []string{"-input", "-", "-metrics.spectrum", "bark"}, want: ErrInvalidValue},
		"SpectrumSize":     {args: []string{"-input", "-", "-metrics.spectrum.size", "1000"}, want: ErrInvalidValue},
		"SmallSpectrum":    {args: []string{"-input", "-", "-metrics.spectrum.size", "2"}, want: ErrInvalidValue},
		"HugeSpectrum":     {args: []string{"-input", "-", "-metrics.spectrum.size", "131072"}, want: ErrInvalidValue},
		"BadLogLevel":      {args: []string{"-input", "-", "-log.level", "loud"}, want: ErrInvalidValue},
		"BadLogFormat":     {args: []string{"-input", "-", "-log.format", "xml"}, want: ErrInvalidValue},
		"MissingFile":      {args: []string{"-input", filepath.Join(t.TempDir(), "missing.wav")}, want: ErrInvalidInput},
		"Directory":        {args: []string{"-input", t.TempDir()}, want: ErrInvalidInput},
		"BadScheme":        {args: []string{"-input", "ftp://host/a.wav"}, want: ErrInvalidInput},
		"NoHost":           {args: []string{"-input", "http:///a.wav"}, want: ErrInvalidInput},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parseConfig(tc.args, env(tc.env), io.Discard)
			require.ErrorIs(t, err, tc.want)
		})
	}

	_, err := parseConfig([]string{"-h"}, env(nil), io.Discard)
	require.True(t, errors.Is(err, flag.ErrHelp))
}

func TestParseInput(t *testing.T) {
	file := filepath.Join(t.TempDir(), "in.wav")
	require.NoError(t, os.WriteFile(file, nil, 0o600))

	in, err := parseInput("-")
	require.NoError(t, err)
	require.True(t, in.stdin)

	in, err = parseInput("https://radio.example/stream.wav")
	require.NoError(t, err)
	require.Equal(t, "https://radio.example/stream.wav", in.url)

	in, err = parseInput(file)
	require.NoError(t, err)
	require.Equal(t, file, in.path)

	in, err = parseInput("file://" + file)
	require.NoError(t, err)
	require.Equal(t, file, in.path)
}
