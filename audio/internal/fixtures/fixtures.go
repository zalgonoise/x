// Package fixtures embeds the module's audio test files and their manifest, so any package's tests (encoding, SDK,
// cmd) can load them by name. Import it only from _test.go files.
//
// Every file under testdata/wav is described in testdata/manifest.json (checksum, format, frames, chunk layout and
// provenance). Regenerate the manifest after adding a file with:
//
//	go test ./internal/fixtures -run TestManifest -update
//
// and fill in its Source, Tool and License by hand (see testdata/PROVENANCE.md).
package fixtures

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"testing"
)

//go:embed testdata
var files embed.FS

// Meta describes one fixture file.
type Meta struct {
	Name          string   `json:"name"`       // path relative to testdata/wav, e.g. "amen_kick/amen_kick_mono_16bit_44100hz.wav"
	SHA256        string   `json:"sha256"`     // of the whole file
	Format        uint16   `json:"format"`     // wFormatTag, or the SubFormat of an extensible file
	Extensible    bool     `json:"extensible"` // WAVE_FORMAT_EXTENSIBLE
	Channels      uint16   `json:"channels"`
	SampleRate    uint32   `json:"sample_rate"`
	BitsPerSample uint16   `json:"bits_per_sample"`
	ValidBits     uint16   `json:"valid_bits"`
	DataSize      uint32   `json:"data_size"`         // declared data chunk size
	Frames        int      `json:"frames"`            // whole frames in the data chunk
	Chunks        []string `json:"chunks"`            // chunk IDs in file order
	DistinctChans bool     `json:"distinct_channels"` // false when every channel carries the same samples (dual mono)
	Source        string   `json:"source"`
	Tool          string   `json:"tool"`
	License       string   `json:"license"`
	Notes         string   `json:"notes,omitempty"`
}

// Fixture is a fixture file and its metadata.
type Fixture struct {
	Meta
	Data []byte
}

// FS returns the embedded testdata/wav tree.
func FS() fs.FS {
	sub, err := fs.Sub(files, "testdata/wav")
	if err != nil {
		panic(err) // the embedded tree always contains testdata/wav
	}

	return sub
}

// Read returns the contents of the fixture with the given name (relative to testdata/wav).
func Read(name string) ([]byte, error) {
	return fs.ReadFile(files, path.Join("testdata/wav", name))
}

// WAV returns the contents of a fixture, failing the test if it cannot be read.
func WAV(tb testing.TB, name string) []byte {
	tb.Helper()

	b, err := Read(name)
	if err != nil {
		tb.Fatalf("fixtures: %v", err)
	}

	return b
}

// Manifest returns the metadata of every fixture, sorted by name.
func Manifest() ([]Meta, error) {
	b, err := files.ReadFile("testdata/manifest.json")
	if err != nil {
		return nil, err
	}

	var metas []Meta
	if err := json.Unmarshal(b, &metas); err != nil {
		return nil, fmt.Errorf("fixtures: decoding manifest: %w", err)
	}

	sort.Slice(metas, func(i, j int) bool { return metas[i].Name < metas[j].Name })

	return metas, nil
}

// All returns every fixture accepted by keep (nil keeps all), failing the test on error.
func All(tb testing.TB, keep func(Meta) bool) []Fixture {
	tb.Helper()

	metas, err := Manifest()
	if err != nil {
		tb.Fatalf("fixtures: %v", err)
	}

	out := make([]Fixture, 0, len(metas))

	for i := range metas {
		if keep != nil && !keep(metas[i]) {
			continue
		}

		out = append(out, Fixture{Meta: metas[i], Data: WAV(tb, metas[i].Name)})
	}

	return out
}
