package window

import (
	"errors"
	"fmt"
)

var blackmanMap = map[int]Block{
	8:    Blackman8,
	16:   Blackman16,
	32:   Blackman32,
	64:   Blackman64,
	128:  Blackman128,
	256:  Blackman256,
	512:  Blackman512,
	1024: Blackman1024,
	2048: Blackman2048,
	4096: Blackman4096,
	8192: Blackman8192,
}

var ErrInvalidBlockSize = errors.New("fft/window: invalid block size")

func Blackman(i int) (Block, error) {
	w, ok := blackmanMap[i]
	if !ok {
		return nil, fmt.Errorf("%w: size %d doesn't have a precomputed window", ErrInvalidBlockSize, i)
	}
	return w, nil
}
