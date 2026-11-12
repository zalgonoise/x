package audio

import (
	"errors"
	"io"
)

// SampleReader is to decoded audio what io.Reader is to bytes. Every format decoder implements it.
type SampleReader interface {
	// Format describes the samples the reader produces.
	Format() Format
	// ReadSamples decodes up to len(dst) interleaved, normalised samples into dst and returns how many were written,
	// always a multiple of Format().Channels (len(dst) should be one too). It returns io.EOF at the end of the stream,
	// possibly together with n > 0.
	ReadSamples(dst []float64) (n int, err error)
}

// SampleWriter is the encoding counterpart of SampleReader.
type SampleWriter interface {
	// Format describes the samples the writer expects.
	Format() Format
	// WriteSamples encodes interleaved, normalised samples and returns how many were consumed. len(src) must be a
	// multiple of Format().Channels.
	WriteSamples(src []float64) (n int, err error)
}

const (
	// readAllSamples is about how many samples ReadAll asks for per call: whole frames, at least one. It does not scale
	// with the channel count, which comes from untrusted headers.
	readAllSamples = 1 << 15
	// maxEmptyReads is how many consecutive (0, nil) reads ReadAll tolerates before giving up with io.ErrNoProgress.
	maxEmptyReads = 100
)

// ReadAll reads every sample from r until io.EOF and returns them as a Buffer. A reader that returns no samples and no
// error 100 times in a row makes it fail with io.ErrNoProgress.
func ReadAll(r SampleReader) (Buffer, error) {
	f := r.Format()
	buf := Buffer{Format: f}
	channels := max(f.Channels, 1)
	chunk := make([]float64, max(readAllSamples/channels, 1)*channels)

	for empty := 0; ; {
		n, err := r.ReadSamples(chunk)
		buf.Samples = append(buf.Samples, chunk[:n]...)

		switch {
		case errors.Is(err, io.EOF):
			return buf, nil
		case err != nil:
			return buf, err
		case n > 0:
			empty = 0
		default:
			empty++

			if empty >= maxEmptyReads {
				return buf, io.ErrNoProgress
			}
		}
	}
}
