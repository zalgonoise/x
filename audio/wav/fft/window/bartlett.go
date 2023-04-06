package window

import (
	"fmt"
)

var bartlettMap = map[int]Block{
	8:    Bartlett8,
	16:   Bartlett16,
	32:   Bartlett32,
	64:   Bartlett64,
	128:  Bartlett128,
	256:  Bartlett256,
	512:  Bartlett512,
	1024: Bartlett1024,
	2048: Bartlett2048,
	4096: Bartlett4096,
	8192: Bartlett8192,
}

func Bartlett(i int) (Block, error) {
	w, ok := bartlettMap[i]
	if !ok {
		return nil, fmt.Errorf("%w: size %d doesn't have a precomputed window", ErrInvalidBlockSize, i)
	}
	return w, nil
}
