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

func MaxSpectra(data [][]fft.FrequencyPower) ([]fft.FrequencyPower, error) {
	if len(data) == 0 {
		return nil, nil
	}

	for i := range data {
		slices.SortFunc(data[i], func(a, b fft.FrequencyPower) int {
			return cmp.Compare(a.Mag, b.Mag)
		})
	}

	final := make([]fft.FrequencyPower, len(data))
	for i := range final {
		if len(data[i]) == 0 {
			continue
		}

		final[i] = data[i][0]
	}

	slices.SortFunc(final, func(a, b fft.FrequencyPower) int {
		return cmp.Compare(a.Mag, b.Mag)
	})

	return final, nil
}
