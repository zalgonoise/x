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

var blockSizeMap = map[int]BlockSize{
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

// AsBlock returns a valid BlockSize for the input int `size`. If the input
// size is not valid, a default BlockSize is returned (Block1024)
func AsBlock(size int) BlockSize {
	if bs, ok := blockSizeMap[size]; ok {
		return bs
	}
	return Block1024
}
