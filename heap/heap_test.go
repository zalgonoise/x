package heap

import (
	_ "embed"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:generate go run github.com/zalgonoise/x/heap/scripts/build-testdata -n 500 -o testdata/int_500.json
//go:generate go run github.com/zalgonoise/x/heap/scripts/build-testdata -n 10000 -o testdata/int_10000.json
//go:generate go run github.com/zalgonoise/x/heap/scripts/build-testdata -n 10000000 -o testdata/int_10000000.json
//go:generate go run github.com/zalgonoise/x/heap/scripts/build-testdata -n 100000000 -o testdata/int_100000000.json

var (
	//go:embed testdata/int_500.json
	int500 []byte

	//go:embed testdata/int_10000.json
	int10000 []byte

	//go:embed testdata/int_10000000.json
	int10000000 []byte

	//go:embed testdata/int_100000000.json
	int100000000 []byte
)

func TestSort(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		data  []int
		wants []int
	}{
		{
			name:  "Success",
			data:  []int{1, 6, 21, 13, 7, 4, 16, 8, 14, 5, 2},
			wants: []int{1, 2, 4, 5, 6, 7, 8, 13, 14, 16, 21},
		},
		{
			name:  "Success/WithRepeatedElements",
			data:  []int{1, 21, 13, 7, 4, 1, 16, 8, 14, 5, 2},
			wants: []int{1, 1, 2, 4, 5, 7, 8, 13, 14, 16, 21},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			Sort(testcase.data)

			require.Equal(t, testcase.data, testcase.wants)
		})
	}
}

func BenchmarkSort(b *testing.B) {
	values, err := loadValues()
	require.NoError(b, err)

	for _, testcase := range []struct {
		name  string
		items []int
	}{
		{
			name:  "500Elements",
			items: values[500],
		},
		{
			name:  "10000Elements",
			items: values[10000],
		},
		{
			name:  "10000000Elements",
			items: values[10000000],
		},
		{
			name:  "100000000Elements",
			items: values[100000000],
		},
	} {
		b.Run(testcase.name, func(b *testing.B) {
			for b.Loop() {
				b.StopTimer()
				s := make([]int, len(testcase.items))
				copy(s, testcase.items)
				b.StartTimer()

				Sort(s)
			}
		})
	}
}

func loadValues() (map[int][]int, error) {
	int500slice := make([]int, 0, 500)
	int10000slice := make([]int, 0, 10000)
	int10000000slice := make([]int, 0, 10000000)
	int100000000slice := make([]int, 0, 100000000)

	errs := make([]error, 0, 4)

	if err := json.Unmarshal(int500, &int500slice); err != nil {
		errs = append(errs, err)
	}

	if err := json.Unmarshal(int10000, &int10000slice); err != nil {
		errs = append(errs, err)
	}
	if err := json.Unmarshal(int10000000, &int10000000slice); err != nil {
		errs = append(errs, err)
	}
	if err := json.Unmarshal(int100000000, &int100000000slice); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return map[int][]int{
		500:       int500slice,
		10000:     int10000slice,
		10000000:  int10000000slice,
		100000000: int100000000slice,
	}, nil
}
