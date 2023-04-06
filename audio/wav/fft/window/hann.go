package window

import (
	"fmt"
)

var hannMap = map[int]Block{
	8:    Hann8,
	16:   Hann16,
	32:   Hann32,
	64:   Hann64,
	128:  Hann128,
	256:  Hann256,
	512:  Hann512,
	1024: Hann1024,
	2048: Hann2048,
	4096: Hann4096,
	8192: Hann8192,
}

func Hann(i int) (Block, error) {
	w, ok := hannMap[i]
	if !ok {
		return nil, fmt.Errorf("%w: size %d doesn't have a precomputed window", ErrInvalidBlockSize, i)
	}
	return w, nil
}
