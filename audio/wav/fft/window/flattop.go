package window

import (
	"fmt"
)

var flattopMap = map[int]Block{
	8:    FlatTop8,
	16:   FlatTop16,
	32:   FlatTop32,
	64:   FlatTop64,
	128:  FlatTop128,
	256:  FlatTop256,
	512:  FlatTop512,
	1024: FlatTop1024,
	2048: FlatTop2048,
	4096: FlatTop4096,
	8192: FlatTop8192,
}

func FlatTop(i int) (Block, error) {
	w, ok := flattopMap[i]
	if !ok {
		return nil, fmt.Errorf("%w: size %d doesn't have a precomputed window", ErrInvalidBlockSize, i)
	}
	return w, nil
}
