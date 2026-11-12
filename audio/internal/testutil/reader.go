package testutil

import (
	"io"
)

// ChunkReader returns a reader that serves r in reads of at most sizes[0], sizes[1], ... bytes, cycling through
// sizes. It reproduces network-like short reads, including reads that split headers and samples. Sizes <= 0 are
// treated as 1.
func ChunkReader(r io.Reader, sizes ...int) io.Reader {
	if len(sizes) == 0 {
		sizes = []int{1}
	}

	return &chunkReader{r: r, sizes: sizes}
}

type chunkReader struct {
	r     io.Reader
	sizes []int
	i     int
}

func (c *chunkReader) Read(p []byte) (int, error) {
	size := max(c.sizes[c.i%len(c.sizes)], 1)
	c.i++

	if len(p) > size {
		p = p[:size]
	}

	return c.r.Read(p)
}

// ReadSizes are read sizes that exercise frame-alignment bugs: odd sizes, sizes smaller than a sample, sizes that are
// not multiples of any block align, and sizes around common network buffer lengths.
var ReadSizes = [][]int{ //nolint:gochecknoglobals // shared test table
	{1},
	{3},
	{5},
	{7},
	{1023},
	{1448}, // typical TCP payload
	{4097},
	{2, 3, 5, 7, 11, 13},
}
