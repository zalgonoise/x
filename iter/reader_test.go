package iter_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/iter"
)

var errReader = errors.New("test reader error")

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errReader
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

func TestReadChunks(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input io.Reader
		size  int
		wants []string
		err   error
		ok    bool
	}{
		{
			name:  "ExactSize",
			input: bytes.NewReader([]byte("chunk")),
			size:  5,
			wants: []string{"chunk"},
			ok:    true,
		},
		{
			name:  "UnderSize",
			input: bytes.NewReader([]byte("chunk")),
			size:  10,
			wants: []string{"chunk"},
			ok:    true,
		},
		{
			name:  "OverSize",
			input: bytes.NewReader([]byte("chunk")),
			size:  2,
			wants: []string{"ch", "un", "k"},
			ok:    true,
		},
		{
			name:  "ReaderError",
			input: errorReader{},
			size:  5,
			wants: []string{},
			err:   errReader,
		},
		{
			name:  "ZeroBytesRead",
			input: zeroReader{},
			size:  5,
			wants: []string{},
			ok:    true,
		},
		{
			name:  "HaltSequence",
			input: bytes.NewReader([]byte("halt!")),
			size:  5,
			wants: []string{},
			ok:    false,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			res := make([]string, 0, len(testcase.wants))
			seq := iter.ReadChunks(testcase.input, testcase.size)

			ok := seq(func(data []byte, err error) bool {
				if err != nil {
					require.ErrorIs(t, err, testcase.err)

					return true
				}

				if string(data) == "halt!" {
					return false
				}

				res = append(res, string(data))

				return true
			})

			require.Equal(t, testcase.ok, ok)
			require.Equal(t, len(testcase.wants), len(res))
			for i := range testcase.wants {
				require.Equal(t, testcase.wants[i], res[i])
			}
		})
	}
}

func FuzzReadChunks(f *testing.F) {
	f.Add([]byte("chunk"), 5)
	f.Add([]byte("a longer string! Wow!!"), 3)

	f.Fuzz(func(t *testing.T, input []byte, size int) {
		seq := iter.ReadChunks(bytes.NewReader(input), size)
		ok := seq(func(data []byte, err error) bool {
			if err != nil {
				t.Fail()

				return false
			}

			return true
		})

		require.True(t, ok)
	})
}

// BenchmakrReadChunks evaluates the performance of the ReadChunks call.
//
// Command:
//
//	go test -benchtime=10s -benchmem -bench '^(BenchmarkReadChunks)$' -run '^$'  -cpuprofile=/tmp/cpu.pprof  ./...
//
// Results:
// 2024-05-26 (initial commit)
//
//	goos: darwin
//	goarch: arm64
//	pkg: github.com/zalgonoise/x/iter
//
// benchmark                    iter      time/iter   bytes alloc         allocs
// ---------                    ----      ---------   -----------         ------
// BenchmarkReadChunks-10   57791439   191.00 ns/op      120 B/op   10 allocs/op
func BenchmarkReadChunks(b *testing.B) {
	input := []byte("a long string that consists of 10 6-byte iteration benchmark")
	fn := func(data []byte, err error) bool {
		if err != nil {
			b.Fail()

			return false
		}

		return true
	}

	for i := 0; i < b.N; i++ {
		if !iter.ReadChunks(bytes.NewReader(input), 6)(fn) {
			b.Fail()

			return
		}
	}
}
