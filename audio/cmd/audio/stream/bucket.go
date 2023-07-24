package stream

import "unsafe"

var frequencyValues = []int{
	125, 250, 500, 1000, 2000, 4000, 6000, 8000, 16000, 22000,
}

var frequencyLabels = []string{
	"125", "250", "500", "1000", "2000", "4000", "6000", "8000", "16000", "22000",
}

type bucketConstraint interface {
	int | uint | float32 | float64
}

type bucketMapper[T bucketConstraint] struct {
	values []T
	labels []string
}

func newBucketMapper[T bucketConstraint](values []T, labels []string) *bucketMapper[T] {
	if len(values) == 0 || len(values) != len(labels) {
		return &bucketMapper[T]{
			values: *(*[]T)(unsafe.Pointer(&frequencyValues)),
			labels: frequencyLabels,
		}
	}

	return &bucketMapper[T]{values, labels}
}

func (m *bucketMapper[T]) Get(value T) string {
	for i := range m.values {
		if value < m.values[i] {
			return m.labels[i]
		}
	}

	return ""
}
