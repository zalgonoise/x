package ptr

import (
	"math/rand"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestNonNil(t *testing.T) {
	var (
		zero  = 0
		one   = 1
		two   = 2
		three = 3
		four  = 4
	)

	for _, testcase := range []struct {
		name  string
		input []*int
		wants []*int
	}{
		{
			name:  "NoValueIsNil",
			input: []*int{&zero, &one, &two, &three, &four},
			wants: []*int{&zero, &one, &two, &three, &four},
		},
		{
			name:  "AllValuesAreNil",
			input: []*int{nil, nil, nil, nil},
		},
		{
			name:  "NoInput",
			input: []*int{},
		},
		{
			name:  "OneNil",
			input: []*int{nil},
		},
		{
			name:  "OneValue",
			input: []*int{&zero},
			wants: []*int{&zero},
		},
		{
			name:  "AlternateValueAndNil",
			input: []*int{&zero, &one, nil, nil, &two, nil, &three, nil, nil, &four, nil},
			wants: []*int{&zero, &one, &two, &three, &four},
		},
		{
			name:  "AlternateValueWithRepeat",
			input: []*int{&zero, &zero, nil, nil, &zero, nil, &zero, nil, nil, &zero, nil},
			wants: []*int{&zero, &zero, &zero, &zero, &zero},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.wants, NonNil(testcase.input...))
		})
	}
}

func TestCompact(t *testing.T) {
	var (
		zero  = 0
		one   = 1
		two   = 2
		three = 3
		four  = 4
	)

	for _, testcase := range []struct {
		name  string
		input []*int
		wants []*int
	}{
		{
			name:  "NoValueIsNil",
			input: []*int{&zero, &one, &two, &three, &four},
			wants: []*int{&zero, &one, &two, &three, &four},
		},
		{
			name:  "AllValuesAreNil",
			input: []*int{nil, nil, nil, nil},
		},
		{
			name:  "NoInput",
			input: []*int{},
		},
		{
			name:  "OneNil",
			input: []*int{nil},
		},
		{
			name:  "OneValue",
			input: []*int{&zero},
			wants: []*int{&zero},
		},
		{
			name:  "AlternateValueAndNil",
			input: []*int{&zero, &one, nil, nil, &two, nil, &three, nil, nil, &four, nil},
			wants: []*int{&zero, &one, &two, &three, &four},
		},
		{
			name:  "AlternateValueWithRepeatSingleValue",
			input: []*int{&zero, &zero, nil, nil, &zero, nil, &zero, nil, nil, &zero, nil},
			wants: []*int{&zero},
		},
		{
			name:  "AlternateValueWithRepeatDoubleValue",
			input: []*int{&one, nil, &zero, &zero, nil, nil, &one, nil, &zero, nil, nil, &one, &one, nil},
			wants: []*int{&one, &zero},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.wants, Compact(testcase.input...))
		})
	}
}

func BenchmarkNilScan(b *testing.B) {
	init := func(size int) []*int {
		slice := make([]*int, size)

		for i := range slice {
			value := rand.Int()
			if value%3 == 0 {
				slice[i] = nil

				continue
			}

			slice[i] = &value
		}

		return slice
	}

	for _, bench := range []struct {
		name  string
		input []*int
	}{
		{
			name:  "64Values",
			input: init(64),
		},
		{
			name:  "4096Values",
			input: init(4096),
		},
		{
			name:  "10_000_000Values",
			input: init(10_000_000),
		},
	} {
		b.Run("NoteValidIndicesFirst/"+bench.name, func(b *testing.B) {
			var output []*int

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				nonNilIndices := make([]int, 0, len(bench.input))

				for idx := range bench.input {
					if bench.input[idx] != nil {
						nonNilIndices = append(nonNilIndices, idx)
					}
				}

				switch len(nonNilIndices) {
				case 0:
					output = nil

					continue
				case len(bench.input):
					output = bench.input

					continue
				}

				output = make([]*int, 0, len(bench.input)-len(nonNilIndices))

				for idx := range nonNilIndices {
					output = append(output, bench.input[nonNilIndices[idx]])
				}
			}

			_ = output
		})

		b.Run("AppendDirectly/"+bench.name, func(b *testing.B) {
			var output []*int

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				output = make([]*int, 0, len(bench.input))

				for idx := range bench.input {
					if bench.input[idx] != nil {
						output = append(output, bench.input[idx])
					}
				}
			}

			_ = output
		})
	}
}

func BenchmarkCompact(b *testing.B) {
	init := func(size int) []*int {
		slice := make([]*int, size)

		for i := range slice {
			value := rand.Int()
			if value%3 == 0 {
				slice[i] = nil

				continue
			}

			slice[i] = &value
		}

		return slice
	}

	for _, bench := range []struct {
		name  string
		input []*int
	}{
		{
			name:  "64Values",
			input: init(64),
		},
		{
			name:  "4096Values",
			input: init(4096),
		},
		{
			name:  "10_000_000Values",
			input: init(10_000_000),
		},
	} {
		b.Run("NoteValidIndicesFirst/"+bench.name, func(b *testing.B) {
			var output []*int

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				nonNilIndices := make([]int, 0, len(bench.input))
				cache := make(map[uintptr]struct{}, len(bench.input))

				for idx := range bench.input {
					if bench.input[idx] != nil {
						nonNilIndices = append(nonNilIndices, idx)
					}
				}

				switch len(nonNilIndices) {
				case 0:
					output = nil

					continue
				case 1:
					output = []*int{bench.input[nonNilIndices[0]]}

					continue
				}

				output = make([]*int, 0, len(bench.input)-len(nonNilIndices))

				for idx := range nonNilIndices {
					ptr := uintptr(unsafe.Pointer(bench.input[nonNilIndices[idx]]))
					if _, ok := cache[ptr]; ok {
						continue
					}

					cache[ptr] = struct{}{}
					output = append(output, bench.input[nonNilIndices[idx]])
				}
			}

			_ = output
		})

		b.Run("AppendDirectly/"+bench.name, func(b *testing.B) {
			var output []*int

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				output = make([]*int, 0, len(bench.input))
				cache := make(map[uintptr]struct{}, len(bench.input))

				for idx := range bench.input {
					if bench.input[idx] != nil {
						ptr := uintptr(unsafe.Pointer(bench.input[idx]))
						if _, ok := cache[ptr]; ok {
							continue
						}

						cache[ptr] = struct{}{}
						output = append(output, bench.input[idx])
					}
				}
			}

			_ = output
		})
	}
}
