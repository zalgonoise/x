// Package audio holds the format-agnostic core of the module: the description of an audio stream (Format), blocks of
// decoded samples (Buffer), the reader and writer interfaces every codec implements (SampleReader, SampleWriter) and a
// registry that picks a decoder by sniffing the input (RegisterFormat, Decode).
//
// Samples are float64 values normalised to [-1, 1): an n-bit integer sample v decodes to v / 2^(n-1), so the most
// negative value is exactly -1 and the most positive one is 1 - 2^-(n-1) (the libsndfile / ffmpeg convention).
// Multichannel samples are interleaved (L0 R0 L1 R1 ...).
//
// This package depends only on the standard library.
package audio
