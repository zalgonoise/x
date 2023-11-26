package fft

const (
	_ = 1 << iota
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
	//nolint:gochecknoglobals // immutable set of supported block sizes, as a cache map
	supportedSizes = map[int]struct{}{
		Block8:    {},
		Block16:   {},
		Block32:   {},
		Block64:   {},
		Block128:  {},
		Block256:  {},
		Block512:  {},
		Block1024: {},
		Block2048: {},
		Block4096: {},
		Block8192: {},
	}

	//nolint:gochecknoglobals // immutable set of supported block sizes, as a slice
	blockSizes = []int{
		Block8, Block16, Block32, Block64,
		Block128, Block256, Block512, Block1024,
		Block2048, Block4096, Block8192,
	}
)

// AsBlock returns a valid block size for the input int `size`. If the input
// size is not valid, a default block size is returned (Block1024).
func AsBlock(size int) int {
	if _, ok := supportedSizes[size]; ok {
		return size
	}

	return Block1024
}

// NearestBlock finds the correct block size that is nearest to the input value.
//
// The logic behind this function is to iterate through the supported block size,
// evaluating if the input size is equal or smaller than the current index.
//
// If it is equal, the current index is returned. If it is smaller, then the previous
// index block size is returned. The iteration continues while size is greater than the
// current index until we drain all supported block sizes.
//
// If the size is too big, the highest supported block size is returned.
func NearestBlock(size int) int {
	for i := 1; i < len(blockSizes); i++ {
		if size == blockSizes[i] {
			return blockSizes[i]
		}

		if size < blockSizes[i] {
			return blockSizes[i-1]
		}
	}

	return blockSizes[len(blockSizes)-1]
}
