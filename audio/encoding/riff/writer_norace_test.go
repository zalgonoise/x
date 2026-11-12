//go:build !race

package riff_test

import (
	"bytes"
	"math"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/riff"
)

// Not built with -race: checkptr rejects the oversized slice header (make cover runs it).
func TestWriteChunkOver4GiB(t *testing.T) {
	// a slice header claiming 4 GiB + 2 bytes; WriteChunk must refuse it before touching the body
	var b [1]byte

	body := unsafe.Slice(&b[0], math.MaxUint32+3)

	var out bytes.Buffer

	w, err := riff.NewWriter(&out, riff.WAVE)
	require.NoError(t, err)
	require.ErrorIs(t, w.WriteChunk(riff.ID("LIST"), body), riff.ErrSize)
	require.Equal(t, 12, out.Len(), "nothing written")
}
