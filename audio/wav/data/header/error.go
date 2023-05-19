package header

import "github.com/zalgonoise/x/audio/errs"

const (
	ErrDomain = errs.Domain("audio/wav/data/header")

	ErrInvalid = errs.Kind("invalid")
	ErrShort   = errs.Kind("short")

	ErrSubChunkHeader = errs.Entity("subchunk header")
	ErrBuffer         = errs.Entity("buffer")
)

var (
	ErrInvalidSubChunkHeader = errs.New(ErrDomain, ErrInvalid, ErrSubChunkHeader)
	ErrShortBuffer           = errs.New(ErrDomain, ErrShort, ErrBuffer)
)
