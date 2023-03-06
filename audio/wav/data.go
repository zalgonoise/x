package wav

import "github.com/zalgonoise/x/audio/wav/data"

// NewChunk is a factory for data.Chunk interfaces.
//
// The data.Chunk are interfaces wrapping different types, based on the
// bit depth `bitDepth` value. These objects will store slices of integers of
// different sizes (int8, int16, int32, and bytes for "junk"), and the
// data.Chunk interface exposes the needed methods to work seamlessly with those
// different data types
//
// Note: I wanted a cleaner approach to this using generics and type constraints,
// but I was getting nowhere meaningful; and ended up breaking at a certain point
// due to the way that Go handles a slice of a type and its conversions to a different type
func NewChunk(bitDepth uint16, subchunk *data.ChunkHeader) data.Chunk {
	if subchunk != nil && string(subchunk.Subchunk2ID[:]) == junkSubchunkIDString {
		bitDepth = 0
	}

	switch bitDepth {
	case 0:
		if subchunk == nil {
			subchunk = data.NewJunkHeader()
		}
		return &data.ChunkJunk{
			ChunkHeader: subchunk,
			Depth:       0,
		}
	case bitDepth8:
		if subchunk == nil {
			subchunk = data.NewDataHeader()
		}
		return &data.Chunk8bit{
			ChunkHeader: subchunk,
			Depth:       bitDepth8,
		}
	case bitDepth16:
		if subchunk == nil {
			subchunk = data.NewDataHeader()
		}
		return &data.Chunk16bit{
			ChunkHeader: subchunk,
			Depth:       bitDepth16,
		}
	case bitDepth24:
		if subchunk == nil {
			subchunk = data.NewDataHeader()
		}
		return &data.Chunk24bit{
			ChunkHeader: subchunk,
			Depth:       bitDepth24,
		}
	case bitDepth32:
		if subchunk == nil {
			subchunk = data.NewDataHeader()
		}
		return &data.Chunk32bit{
			ChunkHeader: subchunk,
			Depth:       bitDepth32,
		}
	default:
		return nil
	}
}
