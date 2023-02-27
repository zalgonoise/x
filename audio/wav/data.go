package wav

import "github.com/zalgonoise/x/audio/wav/data"

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
