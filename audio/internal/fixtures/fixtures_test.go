package fixtures_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io/fs"
	"os"
	"slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

var update = flag.Bool("update", false, "regenerate testdata/manifest.json from the files under testdata/wav")

const unknown = "unknown (ask the author)"

// TestManifest checks that testdata/manifest.json describes every fixture exactly, as computed by the independent
// reference decoder. Run with -update to regenerate it; provenance fields of existing entries are preserved.
func TestManifest(t *testing.T) {
	var names []string

	require.NoError(t, fs.WalkDir(fixtures.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			names = append(names, p)
		}

		return err
	}))

	existing := map[string]fixtures.Meta{}

	if metas, err := fixtures.Manifest(); err == nil {
		for _, m := range metas {
			existing[m.Name] = m
		}
	} else if !*update {
		t.Fatalf("reading manifest: %v", err)
	}

	got := make([]fixtures.Meta, 0, len(names))

	for _, name := range names {
		b := fixtures.WAV(t, name)
		ref, err := testutil.DecodeRef(b)
		require.NoError(t, err, name)

		sum := sha256.Sum256(b)
		meta := fixtures.Meta{
			Name:          name,
			SHA256:        hex.EncodeToString(sum[:]),
			Format:        ref.Format,
			Extensible:    ref.Extensible,
			Channels:      ref.Channels,
			SampleRate:    ref.SampleRate,
			BitsPerSample: ref.BitsPerSample,
			ValidBits:     ref.ValidBits,
			DataSize:      ref.DataSize,
			Frames:        ref.Frames,
			Chunks:        ref.ChunkIDs,
			DistinctChans: distinctChannels(ref),
			Source:        unknown,
			Tool:          unknown,
			License:       unknown,
		}

		if prev, ok := existing[name]; ok {
			meta.Source, meta.Tool, meta.License, meta.Notes = prev.Source, prev.Tool, prev.License, prev.Notes
		}

		got = append(got, meta)
	}

	// Manifest() sorts by name; WalkDir order differs for names that sort between a directory and its contents
	sort.Slice(got, func(i, j int) bool { return got[i].Name < got[j].Name })

	if *update {
		b, err := json.MarshalIndent(got, "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile("testdata/manifest.json", append(b, '\n'), 0o600))

		return
	}

	want, err := fixtures.Manifest()
	require.NoError(t, err)
	require.Equal(t, want, got, "testdata/manifest.json is stale: run go test ./internal/fixtures -run TestManifest -update")
}

func TestAll(t *testing.T) {
	all := fixtures.All(t, nil)
	require.NotEmpty(t, all)

	stereo := fixtures.All(t, func(m fixtures.Meta) bool { return m.Channels == 2 })
	require.NotEmpty(t, stereo)

	for _, f := range stereo {
		require.Equal(t, uint16(2), f.Channels)
		require.NotEmpty(t, f.Data)
	}
}

func distinctChannels(ref testutil.RefFile) bool {
	if ref.Channels < 2 {
		return false
	}

	if ref.Ints != nil {
		for c := 1; c < len(ref.Ints); c++ {
			if !slices.Equal(ref.Ints[0], ref.Ints[c]) {
				return true
			}
		}

		return false
	}

	for c := 1; c < len(ref.Floats); c++ {
		if !slices.Equal(ref.Floats[0], ref.Floats[c]) {
			return true
		}
	}

	return false
}
