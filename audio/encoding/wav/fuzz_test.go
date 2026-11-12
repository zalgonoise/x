package wav_test

import (
	"bytes"
	"errors"
	"math"
	"testing"
	"testing/iotest"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/riff"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

// fuzzSeeds returns every fixture plus builder-made edge cases: metadata chunks around data, odd sizes, fmt
// extensions, live sizes, truncations and a non-RIFF input.
func fuzzSeeds(f *testing.F) [][]byte {
	f.Helper()

	all := fixtures.All(f, nil)
	seeds := make([][]byte, 0, len(all)+10)

	for i := range all {
		seeds = append(seeds, all[i].Data)
	}

	live := uint32(math.MaxUint32)
	samples := [][]int32{{0, 1, -1, 32767, -32768}, {7, -7, 100, -100, 0}}

	for _, w := range []testutil.WAV{
		{Channels: 2, SampleRate: 48000, BitsPerSample: 16},
		{Channels: 2, SampleRate: 48000, BitsPerSample: 16, FmtSize: 18},
		{Channels: 2, SampleRate: 48000, BitsPerSample: 16, DataSize: &live},
		{Channels: 2, SampleRate: 48000, BitsPerSample: 16, Before: []testutil.Chunk{{ID: "LIST", Data: []byte("odd")}}},
		{Channels: 2, SampleRate: 48000, BitsPerSample: 16, After: []testutil.Chunk{{ID: "id3 ", Data: []byte("tag")}}},
		{Channels: 2, SampleRate: 48000, BitsPerSample: 16, NoPad: true, Before: []testutil.Chunk{{ID: "JUNK", Data: []byte("x")}}},
	} {
		seeds = append(seeds, w.EncodeInts(samples))
	}

	full := seeds[len(seeds)-1]
	seeds = append(seeds, full[:20], full[:40], full[:len(full)-3], []byte("not a wav file at all, but long enough to try"),
		// formats only the Decoder handles
		testutil.WAV{
			Format: testutil.FormatExtensible, Channels: 2, SampleRate: 48000, BitsPerSample: 32, ValidBits: 24,
			ChannelMask: 0x3,
		}.EncodeInts([][]int32{{1 << 8, -1 << 8}, {0, 1 << 30}}),
		testutil.WAV{Format: testutil.FormatFloat, Channels: 1, SampleRate: 48000, BitsPerSample: 64}.
			EncodeFloats([][]float64{{0, 1.5, -1, math.NaN()}}),
		testutil.WAV{Format: testutil.FormatALaw, Channels: 1, SampleRate: 8000, BitsPerSample: 8}.Encode([]byte{0xD5, 0x2A}),
	)

	return seeds
}

// FuzzDecode feeds arbitrary bytes to the io.Writer decode path: it must never panic or loop, and a successful decode
// must re-encode to a file that decodes to the same samples.
func FuzzDecode(f *testing.F) {
	for _, seed := range fuzzSeeds(f) {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		w, err := wav.Decode(bytes.Clone(b))
		if err != nil || w.Data == nil {
			return
		}

		again, err := wav.Decode(w.Bytes())
		if err != nil {
			t.Fatalf("re-encoded file does not decode: %v", err)
		}

		if !equalInts(w.Data.Value(), again.Data.Value()) {
			t.Fatalf("decode -> encode -> decode changed the samples")
		}
	})
}

// FuzzFrom feeds arbitrary bytes, served in small reads, to the io.ReaderFrom decode path; it must never panic, loop or
// allocate from lying chunk sizes.
func FuzzFrom(f *testing.F) {
	for _, seed := range fuzzSeeds(f) {
		f.Add(seed, uint8(7))
	}

	f.Fuzz(func(t *testing.T, b []byte, readSize uint8) {
		w, err := wav.From(testutil.ChunkReader(bytes.NewReader(b), int(readSize)))
		if err != nil || w.Data == nil {
			return
		}

		if got := len(w.Data.Float()); got > len(b) {
			t.Fatalf("decoded %d samples from %d input bytes", got, len(b))
		}
	})
}

// FuzzSplitWrite checks that splitting the input across two Writes decodes the same samples as a single Write.
func FuzzSplitWrite(f *testing.F) {
	for _, seed := range fuzzSeeds(f) {
		f.Add(seed, uint16(37))
	}

	f.Fuzz(func(t *testing.T, b []byte, at uint16) {
		whole, err := wav.Decode(bytes.Clone(b))
		if err != nil || whole.Data == nil {
			return
		}

		split := int(at) % (len(b) + 1)
		w := new(wav.Wav)

		// a prefix of a decodable file may only report "need more data"
		if _, err := w.Write(bytes.Clone(b[:split])); err != nil && !errors.Is(err, wav.ErrShortDataBuffer) {
			t.Fatalf("split at %d: first Write failed: %v", split, err)
		}

		if _, err := w.Write(bytes.Clone(b[split:])); err != nil || w.Data == nil {
			t.Fatalf("split at %d: err=%v data=%v (a single Write decoded fine)", split, err, w.Data != nil)
		}

		if !equalInts(whole.Data.Value(), w.Data.Value()) {
			t.Fatalf("split at %d decoded different samples", split)
		}
	})
}

// FuzzStreamReadFrom drives the Stream (ring buffer + processor) path with arbitrary bytes and read sizes.
func FuzzStreamReadFrom(f *testing.F) {
	for _, seed := range fuzzSeeds(f) {
		f.Add(seed, uint8(3))
	}

	f.Fuzz(func(t *testing.T, b []byte, readSize uint8) {
		var frames int

		s := wav.NewStream(func(h *wav.Header, samples []float64) error {
			if len(samples)%int(h.NumChannels) != 0 {
				t.Fatalf("processor got %d samples for %d channels", len(samples), h.NumChannels)
			}

			frames += len(samples)

			return nil
		}, wav.WithSize(64))

		if _, err := s.ReadFrom(iotest.DataErrReader(testutil.ChunkReader(bytes.NewReader(b), int(readSize)))); err != nil {
			return // invalid input: only the absence of panics and over-reads matters
		}

		if frames > len(b) {
			t.Fatalf("processed %d samples from %d input bytes", frames, len(b))
		}
	})
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// FuzzDecoder feeds arbitrary bytes to the streaming Decoder: it must never panic, loop or return more samples than the
// input holds; reading through small reads must decode the same samples; whenever the independent reference decoder
// also accepts the input, both must agree on the samples; and re-encoding what was decoded (with its metadata) must
// decode to the same samples and metadata again.
func FuzzDecoder(f *testing.F) {
	for _, seed := range fuzzSeeds(f) {
		f.Add(seed, uint8(5))
	}

	f.Fuzz(func(t *testing.T, b []byte, readSize uint8) {
		d, err := wav.NewDecoder(bytes.NewReader(b), wav.WithMetadataLimit(1<<10))
		if err != nil {
			return
		}

		got, err := readAllSamples(t, d, 64*d.Format().Channels)
		if err != nil {
			t.Fatalf("a bytes.Reader cannot fail mid-stream: %v", err)
		}

		if len(got)*d.Format().SampleSize() > len(b) {
			t.Fatalf("decoded %d samples from %d input bytes", len(got), len(b))
		}

		split, err := wav.NewDecoder(testutil.ChunkReader(bytes.NewReader(b), int(readSize)%13+1))
		if err != nil {
			t.Fatalf("small reads fail where one read succeeded: %v", err)
		}

		again, err := readAllSamples(t, split, split.Format().Channels)
		if err != nil {
			t.Fatalf("small reads: %v", err)
		}

		requireSameSamples(t, got, again)

		if _, err := testutil.DecodeRef(b); err == nil {
			requireSameSamples(t, refSamples(t, b), got)
		}

		requireTranscodes(t, d, got)
	})
}

// requireTranscodes re-encodes samples decoded by d, with d's format and metadata, into a stream, and checks that it
// decodes to the same format, metadata and samples.
func requireTranscodes(t *testing.T, d *wav.Decoder, samples []float64) {
	t.Helper()

	var withData, kept []riff.Chunk

	for _, c := range d.Chunks() {
		if c.Data == nil {
			continue // listed without data (metadata limit): it cannot be written
		}

		withData = append(withData, c)

		if c.ID != riff.ID("fact") {
			kept = append(kept, c) // the Encoder drops fact chunks and writes its own
		}
	}

	var buf bytes.Buffer

	e, err := wav.NewEncoder(&buf, d.Format(), wav.WithChunks(withData...))
	if err != nil {
		t.Fatalf("cannot encode the decoded format %s: %v", d.Format(), err)
	}

	if _, err := e.WriteSamples(samples); err != nil {
		t.Fatalf("encoding: %v", err)
	}

	if err := e.Close(); err != nil {
		t.Fatalf("closing: %v", err)
	}

	again, err := wav.NewDecoder(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-encoded stream does not decode: %v", err)
	}

	if again.Format() != d.Format() {
		t.Fatalf("format changed from %s to %s", d.Format(), again.Format())
	}

	got, err := readAllSamples(t, again, 64*again.Format().Channels)
	if err != nil {
		t.Fatalf("decoding the re-encoded stream: %v", err)
	}

	requireSameSamples(t, samples, got)

	meta := again.Chunks()
	if again.Format().Encoding != audio.EncodingPCM {
		if len(meta) == 0 || meta[0].ID != riff.ID("fact") {
			t.Fatalf("no fact chunk for %s: %v", again.Format(), meta)
		}

		meta = meta[1:] // written by the Encoder
	}

	if !equalChunks(kept, meta) {
		t.Fatalf("metadata changed: %v -> %v", kept, meta)
	}
}

func equalChunks(a, b []riff.Chunk) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].ID != b[i].ID || !bytes.Equal(a[i].Data, b[i].Data) {
			return false
		}
	}

	return true
}
