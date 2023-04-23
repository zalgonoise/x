package window

import (
	"math"
)

const (
	tau    = math.Pi * 2
	twoTau = math.Pi * 4
)

// WindowFunc is a type of function that generates a Window based on an input size
type WindowFunc func(int) Window

// Generate calls the WindowFunc itself, with the input int `size`
func (w WindowFunc) Generate(size int) Window {
	return w(size)
}

// Window is a custom type for a slice of float64, that repesents a Window that shapes
// input signal
type Window []float64

// Apply will go through the input slice of float64 values, and apply the window to it by
// multiplying the value by the ratio in the respective element in the Window
func (w Window) Apply(value []float64) {
	for idx := range value {
		value[idx] *= w[idx]
	}
}

var blockSizes = []int{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192}

// New creates a new Window from the input WindowFunc and size
//
// The size should be a power of two supported in the fft package's BlockSize.
//
// In case the input size `int` is not a power of two, it will fallback to the next
// available block size, with a series of bit-shifting operations on the input size
// to derive the closest element in the (private) blockSizes slice of ints.
func New(w WindowFunc, size int) Window {
	blockSize := size >> 4
	var count int
	for i := blockSize; i > 0; i = i >> 1 {
		count++
	}
	return w(blockSizes[count])
}
