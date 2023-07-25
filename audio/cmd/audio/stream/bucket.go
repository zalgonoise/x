package stream

import "unsafe"

var frequencyValues = []int{
	125, 250, 500, 1000, 2000, 4000, 6000, 8000, 16000, 22000,
}

var frequencyLabels = []string{
	"125", "250", "500", "1000", "2000", "4000", "6000", "8000", "16000", "22000",
}

type bucketConstraint interface {
	int | uint | float32 | float64 | string
}

type bucketMapper[K bucketConstraint, V comparable] struct {
	values []K
	labels []V
}

func newBucketMapper[K bucketConstraint, V comparable](values []K, labels []V) *bucketMapper[K, V] {
	if len(values) == 0 || len(values) != len(labels) {
		return &bucketMapper[K, V]{
			values: *(*[]K)(unsafe.Pointer(&frequencyValues)),
			labels: *(*[]V)(unsafe.Pointer(&frequencyLabels)),
		}
	}

	return &bucketMapper[K, V]{values, labels}
}

func (m *bucketMapper[K, V]) Get(value K) V {
	for i := range m.values {
		if value < m.values[i] {
			return m.labels[i]
		}
	}

	return *new(V)
}
