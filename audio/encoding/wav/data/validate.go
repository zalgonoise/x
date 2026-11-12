package data

import (
	"github.com/zalgonoise/valigator"
	"github.com/zalgonoise/x/errs"
)

const (
	ErrDomain = errs.Domain("audio/wav/data")

	ErrEmpty   = errs.Kind("empty")
	ErrInvalid = errs.Kind("invalid")
	ErrMissing = errs.Kind("missing")
	ErrShort   = errs.Kind("short")

	ErrSubChunkHeader = errs.Entity("subchunk header")
	ErrHeader         = errs.Entity("header")
	ErrBuffer         = errs.Entity("buffer")
)

var (
	ErrInvalidSubChunkHeader = errs.WithDomain(ErrDomain, ErrInvalid, ErrSubChunkHeader)
	ErrEmptySubChunkHeader   = errs.WithDomain(ErrDomain, ErrEmpty, ErrSubChunkHeader)
	ErrMissingHeader         = errs.WithDomain(ErrDomain, ErrMissing, ErrHeader)
	ErrShortBuffer           = errs.WithDomain(ErrDomain, ErrShort, ErrBuffer)
)

//nolint:gochecknoglobals // is initialized in the background for performance during runtime, as it is never mutated.
var headerValidator = valigator.New(validateHeaderSubChunkID)

// validateHeaderSubChunkID accepts any RIFF FOURCC: four printable ASCII characters (space padded). Chunks other than
// "data" (LIST, fact, JUNK, bext, cue, ...) are kept as opaque chunks.
func validateHeaderSubChunkID(h *Header) error {
	for _, c := range h.Subchunk2ID {
		if c < ' ' || c > '~' {
			return ErrInvalidSubChunkHeader
		}
	}

	return nil
}

// Validate verifies that the input Header `h` is not nil and that it is valid.
func Validate(h *Header) error {
	if h == nil {
		return ErrEmptySubChunkHeader
	}

	return headerValidator.Validate(h)
}
