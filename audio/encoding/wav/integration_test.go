//go:build integration

package wav_test

import (
	"context"
	"errors"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/encoding/wav/data"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

// TestIntegrationStreamOverHTTP decodes every fixture served over a real HTTP connection with network-like framing,
// the way cmd/audio consumes live streams, and checks the processor sees exactly the file's samples in whole frames.
func TestIntegrationStreamOverHTTP(t *testing.T) {
	for _, fx := range fixtures.All(t, nil) {
		want := dataPayload(t, fx.Data)

		for name, opts := range map[string][]testutil.ServeOption{
			"ContentLength": nil,
			"Chunked1000":   {testutil.Chunked(1000)},
			"Chunked1":      {testutil.Chunked(1)},
			"Chunked1447":   {testutil.Chunked(1447)},
		} {
			t.Run(fx.Name+"/"+name, func(t *testing.T) {
				srv := testutil.NewWAVServer(t, fx.Data, opts...)

				res, err := http.Get(srv.URL)
				require.NoError(t, err)

				t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

				var got []float64

				s := wav.NewStream(func(h *wav.Header, samples []float64) error {
					require.Zero(t, len(samples)%int(h.NumChannels))
					got = append(got, samples...)

					return nil
				}, wav.WithSize(4096))

				_, err = s.ReadFrom(res.Body)
				require.NoError(t, err)

				ring, ok := s.Data.(*data.Ring)
				require.True(t, ok)
				require.Equal(t, want, ring.Converter.Bytes(got))
			})
		}
	}
}

// TestIntegrationEndlessLiveStream reads a live stream (data size 0xFFFFFFFF, payload repeated forever) until the
// client gives up, and checks the decoded prefix repeats the file's samples without misalignment.
func TestIntegrationEndlessLiveStream(t *testing.T) {
	file := fixtures.WAV(t, "amen_kick/amen_kick_stereo_24bit_44100hz.wav")
	want := dataPayload(t, file)
	srv := testutil.NewWAVServer(t, file, testutil.Endless(math.MaxUint32), testutil.Chunked(1447))

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, http.NoBody)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

	var got []float64

	s := wav.NewStream(func(_ *wav.Header, samples []float64) error {
		got = append(got, samples...)

		return nil
	}, wav.WithSize(8192))

	_, err = s.ReadFrom(res.Body)
	require.True(t, err == nil || errors.Is(err, context.DeadlineExceeded), "unexpected error: %v", err)

	ring, ok := s.Data.(*data.Ring)
	require.True(t, ok)

	encoded := ring.Converter.Bytes(got)
	require.Greater(t, len(encoded), 2*len(want), "the stream kept going past one copy of the file")

	for off := 0; off+len(want) <= len(encoded); off += len(want) {
		require.Equal(t, want, encoded[off:off+len(want)], "copy at byte %d", off)
	}
}
