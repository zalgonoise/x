package header

func Validate(header *Header) error {
	switch string(header.Subchunk2ID[:]) {
	case JunkIDString, DataIDString:
		return nil
	default:
		return ErrInvalidSubChunkHeader
	}
}
