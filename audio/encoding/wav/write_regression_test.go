package wav_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

// TestWriteReusedBuffer feeds a file through one reused buffer, like a read loop does: Write must copy what it keeps
// (partial frames and headers carried to the next Write), or the caller's next read corrupts it.
func TestWriteReusedBuffer(t *testing.T) {
	file := fixtures.WAV(t, "amen_kick/amen_kick_stereo_24bit_44100hz.wav")
	want := dataPayload(t, file)

	for _, size := range []int{1, 7, 1001, 4096} {
		w := new(wav.Wav)
		buf := make([]byte, size)
		r := bytes.NewReader(file)

		for {
			n, err := r.Read(buf)
			if n > 0 {
				if _, werr := w.Write(buf[:n]); werr != nil {
					require.ErrorIs(t, werr, wav.ErrShortDataBuffer)
				}

				for i := range buf { // scribble over the buffer, as a reuse would
					buf[i] = 0xAA
				}
			}

			if err == io.EOF {
				break
			}
		}

		require.Equal(t, want, w.Data.Bytes(), "buffer size %d", size)
	}
}

// TestWriteRawPCMAfterNew checks that a Wav built with New accepts raw PCM through Write (its data chunk has no
// declared size, so it is unbounded).
func TestWriteRawPCMAfterNew(t *testing.T) {
	w, err := wav.New(44100, 16, 1, wav.PCMFormat)
	require.NoError(t, err)

	_, err = w.Write([]byte{1, 0, 0xFF, 0x7F, 0, 0x80})
	require.NoError(t, err)
	require.Equal(t, []int{1, 32767, -32768}, w.Data.Value())
}

// TestDecodeTrailingBytesAfterData checks that bytes after the data chunk which are not RIFF chunks (zero padding, a
// raw ID3 tag) end the RIFF content instead of failing the decode.
func TestDecodeTrailingBytesAfterData(t *testing.T) {
	file := fixtures.WAV(t, "amen_kick/amen_kick_mono_16bit_44100hz.wav")
	want := dataPayload(t, file)

	for name, tail := range map[string][]byte{
		"NULPadding": make([]byte, 16),
		"ID3v2":      append([]byte("ID3\x04\x00\x00\x00\x00\x00\x10"), make([]byte, 16)...),
	} {
		t.Run(name, func(t *testing.T) {
			w, err := wav.Decode(append(bytes.Clone(file), tail...))
			require.NoError(t, err)
			require.Equal(t, want, w.Data.Bytes())

			s := wav.NewStream(func(*wav.Header, []float64) error { return nil })
			_, err = s.Write(append(bytes.Clone(file), tail...))
			require.NoError(t, err)

			_, err = s.Write([]byte("more trailing bytes"))
			require.NoError(t, err, "later input after the RIFF content is ignored")
		})
	}
}

// TestWriteRejectsNonWAVAtOnce checks that input which is not a WAV header fails immediately (it used to report
// "short data buffer" and keep buffering), and that the error sticks.
func TestWriteRejectsNonWAVAtOnce(t *testing.T) {
	text := bytes.Repeat([]byte("this is not a wav file, just some text "), 10)

	_, err := wav.Decode(text)
	require.ErrorIs(t, err, wav.ErrInvalidHeader)

	s := wav.NewStream(func(*wav.Header, []float64) error { return nil })

	_, err = s.Write(text)
	require.ErrorIs(t, err, wav.ErrInvalidHeader)

	_, err = s.Write(text)
	require.ErrorIs(t, err, wav.ErrInvalidHeader, "the error is latched")

	// a valid header whose fmt chunk claims ~4 GiB (patched in: the builder would really write that much)
	huge := testutil.WAV{Channels: 1, SampleRate: 44100, BitsPerSample: 16}.Encode(nil)
	binary.LittleEndian.PutUint32(huge[16:20], math.MaxUint32-255)

	_, err = wav.Decode(huge)
	require.ErrorIs(t, err, wav.ErrInvalidHeader, "a huge fmt size is rejected instead of awaited")
}

// TestSnapshotBetweenWrites checks that encoding (Bytes) between Writes does not discard bytes carried for the next
// Write.
func TestSnapshotBetweenWrites(t *testing.T) {
	file := fixtures.WAV(t, "amen_kick/amen_kick_stereo_24bit_44100hz.wav")
	want := dataPayload(t, file)

	w := new(wav.Wav)

	for i := 0; i < len(file); i += 1001 {
		if _, err := w.Write(file[i:min(i+1001, len(file))]); err != nil {
			require.ErrorIs(t, err, wav.ErrShortDataBuffer)
		}

		if w.Data != nil {
			_ = w.Bytes()
		}
	}

	require.Equal(t, want, w.Data.Bytes())
}

// TestStreamEncodeOddData checks that the Stream encoder pads an odd-sized data chunk (8-bit mono, odd ring size).
func TestStreamEncodeOddData(t *testing.T) {
	file := testutil.WAV{
		Channels: 1, SampleRate: 44100, BitsPerSample: 8,
		After: []testutil.Chunk{{ID: "LIST", Data: []byte("INFOxxxx")}},
	}.EncodeInts([][]int32{testutil.IntExtremes(8)})

	s := wav.NewStream(func(*wav.Header, []float64) error { return nil }, wav.WithSize(65))

	_, err := s.Write(file)
	require.NoError(t, err)
	require.Equal(t, 1, s.Size%2, "the ring size is odd")

	ref, err := testutil.DecodeRef(s.Bytes())
	require.NoError(t, err)
	require.Equal(t, []string{"fmt ", "data", "LIST"}, ref.ChunkIDs, "the chunk after the odd data chunk stays aligned")
}
