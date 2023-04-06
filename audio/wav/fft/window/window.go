package window

type Block []float64

func (b Block) Apply(v []float64) {
	for i := range v {
		v[i] *= b[i]
	}
}

func (b Block) Len() int {
	return len(b)
}
