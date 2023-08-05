package fft

// BlockSize is an enumeration for FFT BlockSize values, which are a power of 2,
// from 8 to 8192
type BlockSize int

const (
	_ BlockSize = 1 << iota
	_
	_
	Block8
	Block16
	Block32
	Block64
	Block128
	Block256
	Block512
	Block1024
	Block2048
	Block4096
	Block8192
)

var (
	blockSizeMap = map[int]BlockSize{
		8:    Block8,
		16:   Block16,
		32:   Block32,
		64:   Block64,
		128:  Block128,
		256:  Block256,
		512:  Block512,
		1024: Block1024,
		2048: Block2048,
		4096: Block4096,
		8192: Block8192,
	}

	blockSizeInts = []int{
		8, 16, 32, 64, 128, 256,
		512, 1024, 2048, 4096, 8192,
	}

	blockSizes = []BlockSize{
		Block8, Block16, Block32, Block64,
		Block128, Block256, Block512,
		Block1024, Block2048, Block4096,
		Block8192,
	}
)

// AsBlock returns a valid BlockSize for the input int `size`. If the input
// size is not valid, a default BlockSize is returned (Block1024)
func AsBlock(size int) BlockSize {
	if bs, ok := blockSizeMap[size]; ok {
		return bs
	}
	return Block1024
}

// NearestBlock finds the correct BlockSize that is nearest to the input value.
//
// The logic behind this function is to iterate through the supported BlockSize,
// evaluating if the input size is equal or smaller than the current index.
//
// If it is equal, the current index is returned. If it is smaller, then the previous
// index BlockSize is returned. The iteration continues while size is greater than the
// current index until we drain all supported BlockSize.
//
// If the size is too big, the highest supported BlockSize is returned.
func NearestBlock(size int) BlockSize {
	for i := 1; i < len(blockSizeInts); i++ {
		if size == blockSizeInts[i] {
			return blockSizes[i]
		}

		if size < blockSizeInts[i] {
			return blockSizes[i-1]
		}
	}

	return Block8192
}
