package header

import "github.com/zalgonoise/x/audio/validation"

var headerValidator = validation.New[*Header](validateHeaderSubChunkID)

func validateHeaderSubChunkID(h *Header) error {
	switch string(h.Subchunk2ID[:]) {
	case JunkIDString, DataIDString:
		return nil
	default:
		return ErrInvalidSubChunkHeader
	}
}

// Validate verifies that the input Header `h` is not nil and that it is valid
func Validate(h *Header) error {
	if h == nil {
		return ErrEmptySubChunkHeader
	}

	return headerValidator.Validate(h)
}
