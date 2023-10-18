package stream

// WithFilter appends the input slice of StreamFilter `fns` to the WavBuffer Filters,
// returning the same WavBuffer to allow chaining
func (w *Wav) WithFilter(fns ...StreamFilter) *Wav {
	for _, fn := range fns {
		if fn != nil {
			w.Filters = append(w.Filters, fn)
		}
	}
	return w
}

// Ratio sets the ring buffer's size ratio, short-circuiting if the input
// float64 `ratio` is zero; returning the same WavBuffer to allow chaining
func (w *Wav) Ratio(ratio float64) *Wav {
	if ratio == 0 {
		return w
	}
	w.ratio = ratio
	return w
}

func (w *Wav) BlockSize(size int) *Wav {
	if size < 0 {
		size = 0
	}
	w.blockSize = size
	return w
}
