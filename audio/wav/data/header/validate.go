package header

import "github.com/zalgonoise/x/audio/validation"

var subChunkIDValidator = validation.New[string](
	ErrInvalidSubChunkHeader,
	JunkIDString,
	DataIDString,
)

func Validate(header *Header) error {
	return subChunkIDValidator.Validate(string(header.Subchunk2ID[:]))
}
