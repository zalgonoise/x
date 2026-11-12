//go:build integration

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/internal/fixtures"
)

// TestIntegrationBinaryOnFile builds cmd/audio and runs it on a WAV file: it exits 0 as soon as the input ends, having
// written the per-channel levels to stdout and its logs to stderr.
func TestIntegrationBinaryOnFile(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "audio")

	build := exec.Command("go", "build", "-o", bin, ".")
	build.Stderr = os.Stderr
	require.NoError(t, build.Run())

	input := filepath.Join(dir, "in.wav")
	require.NoError(t, os.WriteFile(input, fixtures.WAV(t, "amen_kick/amen_kick_stereo_16bit_44100hz.wav"), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, bin, "-input", input, "-duration", "1m")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")} // no AUDIO_* from the environment
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()

	require.NoError(t, cmd.Run(), stderr.String())
	require.Less(t, time.Since(start), 10*time.Second, "exits when the file ends, not after -duration")

	require.Contains(t, stderr.String(), "input ended")
	require.False(t, strings.Contains(stderr.String(), "level=ERROR"), stderr.String())
	require.Contains(t, stdout.String(), `"msg":"audio.peak"`)
	require.Contains(t, stdout.String(), `"channel":1`)
}
