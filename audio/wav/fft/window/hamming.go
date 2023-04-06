package window

import (
	"fmt"
)

var hammingMap = map[int]Block{
	8:    Hamming8,
	16:   Hamming16,
	32:   Hamming32,
	64:   Hamming64,
	128:  Hamming128,
	256:  Hamming256,
	512:  Hamming512,
	1024: Hamming1024,
	2048: Hamming2048,
	4096: Hamming4096,
	8192: Hamming8192,
}

func Hamming(i int) (Block, error) {
	w, ok := hammingMap[i]
	if !ok {
		return nil, fmt.Errorf("%w: size %d doesn't have a precomputed window", ErrInvalidBlockSize, i)
	}
	return w, nil
}
