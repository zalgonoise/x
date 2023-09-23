package compactors

import (
	"cmp"
	"slices"

	"github.com/zalgonoise/x/audio/errs"
	"github.com/zalgonoise/x/audio/fft"
)

const (
	errDomain = errs.Domain("x/audio/sdk/compactors")

	ErrEmpty = errs.Kind("empty")

	ErrValueSet = errs.Entity("set of values")
)

var (
	ErrEmptyValueSet = errs.New(errDomain, ErrEmpty, ErrValueSet)
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Max finds the biggest (ordered) value in a slice of a given type, with a bigger-than approach,
// meaning that it will work for positive integer and float values.
//
// To include negative values, an AbsMax approach would be required.
func Max[T cmp.Ordered](values []T) (T, error) {
	if len(values) == 0 {
		return *new(T), ErrEmptyValueSet
	}

	maximum := values[0]

	for i := 1; i < len(values); i++ {
		if values[i] > maximum {
			maximum = values[i]
		}
	}

	return maximum, nil
}

// AbsMax finds the absolute maximum Number value in the input slice.
//
// Using and absolute-approach, the function will return the biggest value in the slice, regardless if
// it is positive or negative. For example, calling AbsMax on the following slice would return the
// negative value of -1.9:  []float64{0.1, 0.9, -0.3, -1.9}.
//
// AbsMax may be slower than Max, and also it will not be compatible with its support for uintptr and string types.
func AbsMax[T Number](values []T) (T, error) {
	if len(values) == 0 {
		return *new(T), ErrEmptyValueSet
	}

	var zero T
	idx := 0
	maximum := values[0]

	for i := 0; i < len(values); i++ {
		if values[i] >= zero {
			if values[i] > maximum {
				maximum = values[i]
				idx = i
			}

			continue
		}

		absValue := -values[i]
		if absValue > maximum {
			maximum = absValue
			idx = i
		}
	}

	return values[idx], nil
}

// MaxSpectra reduces a matrix of frequencies (several registries of sets of frequencies), into a single (ordered) set.
//
// Its strategy involves sorting each set in the matrix to present the strongest magnitude frequencies as the first
// element, and collects these into a new slice (of the same capacity as the matrix).
//
// Finally, it sorts the final slice once again, putting the strongest magnitude frequencies at the beginning of the
// slice, so consumers can consume it up straight away.
func MaxSpectra(data [][]fft.FrequencyPower) ([]fft.FrequencyPower, error) {
	if len(data) == 0 {
		return nil, nil
	}

	for i := range data {
		slices.SortFunc(data[i], func(a, b fft.FrequencyPower) int {
			return cmp.Compare(b.Mag, a.Mag)
		})
	}

	final := make([]fft.FrequencyPower, 0, len(data))
	for i := range data {
		if len(data[i]) > 0 {
			final = append(final, data[i][0])
		}
	}

	slices.SortFunc(final, func(a, b fft.FrequencyPower) int {
		return cmp.Compare(b.Mag, a.Mag)
	})

	return final, nil
}
