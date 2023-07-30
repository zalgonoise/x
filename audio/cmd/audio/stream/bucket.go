package stream

var frequencyValues = []int{
	125, 250, 500, 1000, 2000, 4000, 6000, 8000, 16000, 22000,
}

var frequencyLabels = []string{
	"125", "250", "500", "1000", "2000", "4000", "6000", "8000", "16000", "22000",
}

type bucket[T any] struct {
	values []T
	labels []string
	less   LessFunc[T]
}

func newBucket[T any](values []T, labels []string, lessFunc LessFunc[T]) *bucket[T] {
	if len(values) == 0 || len(values) != len(labels) || lessFunc == nil {
		return nil
	}

	return &bucket[T]{values, labels, lessFunc}
}

func (m bucket[T]) Get(value T) (label string) {
	for i := range m.values {
		if m.less(value, m.values[i]) {
			return m.labels[i]
		}
	}

	return ""
}
