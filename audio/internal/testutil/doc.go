// Package testutil holds test helpers shared across the module's packages.
//
// It must only be imported from _test.go files. Its helpers are deliberately independent from the code under test:
// the WAV builder and reference decoder use encoding/binary only (never encoding/wav or its converters), and the DFT
// oracle is the O(N²) textbook definition. This lets tests compare the library against a second implementation
// instead of against its own output.
//
// Contents:
//   - signals: Sine, Cosine, Impulse, DC, Alternating, Noise, Chirp, Interleave, Deinterleave, IntExtremes
//   - tolerances: ULPDiff, RequireULPs, RequireInDelta, DB, RequireDB
//   - oracles: NaiveDFT, Magnitudes
//   - WAV files: WAV (builder), Chunk, DecodeRef (reference RIFF/WAVE decoder)
//   - I/O: ChunkReader and the ReadSizes table (network-like short reads); combine with testing/iotest
//   - HTTP: NewWAVServer with Chunked, FlushEvery, Endless, DisconnectAfter, Status, ContentType and Delay options;
//     LiveHeader rewrites a file's data size to 0 or 0xFFFFFFFF like live streaming servers do
package testutil
