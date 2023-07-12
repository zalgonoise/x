package wav_test

import (
	"bytes"
	"embed"
	_ "embed"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/data/filters"
	dataheader "github.com/zalgonoise/x/audio/wav/data/header"
	"github.com/zalgonoise/x/audio/wav/header"
)

//go:embed data/internal/testdata/amen_kick/*
var testdataFS embed.FS

type testdata struct {
	data        []byte
	bitDepth    int
	sampleRate  int
	numChannels int
}

func load() ([]testdata, error) {
	var (
		err              error
		mono8bit44100    []byte
		mono16bit44100   []byte
		mono24bit44100   []byte
		mono32bit44100   []byte
		mono32bit96000   []byte
		mono32bit192000  []byte
		mono8bit176400   []byte
		stereo8bit44100  []byte
		stereo16bit44100 []byte
		stereo24bit44100 []byte
		stereo32bit44100 []byte
	)

	if mono8bit44100, err = testdataFS.ReadFile("data/internal/testdata/amen_kick/amen_kick_mono_8bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if mono16bit44100, err = testdataFS.ReadFile("data/internal/testdata/amen_kick/amen_kick_mono_16bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if mono24bit44100, err = testdataFS.ReadFile("data/internal/testdata/amen_kick/amen_kick_mono_24bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if mono32bit44100, err = testdataFS.ReadFile("data/internal/testdata/amen_kick/amen_kick_mono_32bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if mono32bit96000, err = testdataFS.ReadFile("data/internal/testdata/amen_kick/amen_kick_mono_32bit_96000hz.wav"); err != nil {
		return nil, err
	}
	if mono32bit192000, err = testdataFS.ReadFile("data/internal/testdata/amen_kick/amen_kick_mono_32bit_192000hz.wav"); err != nil {
		return nil, err
	}
	if mono8bit176400, err = testdataFS.ReadFile("data/internal/testdata/amen_kick/amen_kick_mono_8bit_176400hz.wav"); err != nil {
		return nil, err
	}
	if stereo8bit44100, err = testdataFS.ReadFile("data/internal/testdata/amen_kick/amen_kick_stereo_8bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if stereo16bit44100, err = testdataFS.ReadFile("data/internal/testdata/amen_kick/amen_kick_stereo_16bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if stereo24bit44100, err = testdataFS.ReadFile("data/internal/testdata/amen_kick/amen_kick_stereo_24bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if stereo32bit44100, err = testdataFS.ReadFile("data/internal/testdata/amen_kick/amen_kick_stereo_32bit_44100hz.wav"); err != nil {
		return nil, err
	}

	_ = mono8bit44100
	_ = mono16bit44100
	_ = mono24bit44100
	_ = mono32bit44100
	_ = mono32bit96000
	_ = mono32bit192000
	_ = mono8bit176400
	_ = stereo8bit44100
	_ = stereo16bit44100
	_ = stereo24bit44100
	_ = stereo32bit44100

	return []testdata{
		{mono8bit44100, 8, 44100, 1},
		{mono16bit44100, 16, 44100, 1},
		{mono24bit44100, 24, 44100, 1},
		{mono32bit44100, 32, 44100, 1},
		{mono32bit96000, 32, 96000, 1},
		{mono32bit192000, 32, 192000, 1},
		{mono8bit176400[:len(mono8bit176400)-1], 8, 176400, 1}, // remove a useless nullbyte in the end
		{stereo8bit44100, 8, 44100, 2},
		{stereo16bit44100, 16, 44100, 2},
		{stereo24bit44100, 24, 44100, 2},
		{stereo32bit44100, 32, 44100, 2},
	}, nil
}

func BenchmarkWav(b *testing.B) {
	td, err := load()
	if err != nil {
		b.Error(err)
		return
	}

	for _, tc := range td[:4] { // mono 44.1kHz 8bit to 32bit
		var loadedWav *wav.Wav

		loadedWav, err = wav.Decode(tc.data)
		if err != nil {
			b.Error(err)
			return
		}

		b.Run(
			"Decode", func(b *testing.B) {
				var (
					w   *wav.Wav
					err error
				)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					w, err = wav.Decode(tc.data)
					if err != nil {
						b.Error(err)
					}
				}
				_ = w
			},
		)
		b.Run(
			"Encode", func(b *testing.B) {
				var buf []byte

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					buf = loadedWav.Bytes()
				}
				_ = buf
			},
		)

		b.Run(
			"Write", func(b *testing.B) {
				var (
					w   *wav.Wav
					err error
				)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					w = new(wav.Wav)
					_, err = w.Write(tc.data)
					if err != nil {
						b.Error(err)
						return
					}
				}
				_ = w
			},
		)

		b.Run(
			"Read", func(b *testing.B) {
				var buf []byte

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					buf = make([]byte, len(tc.data))
					_, err = loadedWav.Read(buf)
					if err != nil {
						b.Error(err)
						return
					}
				}
				_ = buf
			},
		)
	}
}

func TestNewWav(t *testing.T) {
	wants := []byte{
		82, 73, 70, 70, // ChunkID
		0, 0, 0, 0, // ChunkSize
		87, 65, 86, 69, // Format
		102, 109, 116, 32, // Subchunk1ID
		16, 0, 0, 0, // Subchunk1Size
		1, 0, // AudioFormat
		2, 0, // NumChannels
		68, 172, 0, 0, // SampleRate
		16, 177, 2, 0, // ByteRate
		4, 0, // BlockAlign
		16, 0, // BitsPerSample
	}
	w, err := wav.New(44100, 16, 2, 1)
	if err != nil {
		t.Error(err)
	}
	if string(wants) != string(w.Header.Bytes()) {
		t.Errorf("output mismatch error: \n\nwanted %v ;\n\ngot %v\n", wants, w.Header.Bytes())
	}

	parsedHeader, err := header.From(w.Header.Bytes())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(parsedHeader, w.Header) {
		t.Errorf("output mismatch error: \n\nwanted %v ;\n\ngot %v\n", w.Header, parsedHeader)
	}
}

func TestWavDecodeEncode(t *testing.T) {
	testdata, err := load()
	if err != nil {
		t.Error(err)
		return
	}

	for idx, test := range testdata {
		w, err := wav.Decode(test.data)
		if err != nil {
			t.Errorf("decoding error on index %d: %v", idx, err)
			continue
		}
		buf := w.Bytes()
		if len(buf) != len(test.data) {
			t.Errorf("output length mismatch error on index %d: wanted %d ; got %d", idx, len(test.data), len(buf))

		}
		for i := 0; i < len(buf); i++ {
			if buf[i] != test.data[i] {
				t.Errorf("byte mismatch on index %d: #%d wanted %d; got %d -- total len: %d", idx, i, test.data[i], buf[i], len(buf))
				break
			}
		}

		newWav, err := wav.Decode(buf)
		if err != nil {
			t.Errorf("2nd-pass decoding error on index %d: %v", idx, err)
			continue
		}
		newBuf := newWav.Bytes()

		cmp := bytes.Compare(buf, newBuf)
		if cmp != 0 {
			t.Errorf("2nd-pass encoding mismatches the 1st-pass encoding: compare 1st-2nd: %d", cmp)
			continue
		}
		if !reflect.DeepEqual(w, newWav) {
			t.Errorf("output object mismatch error: wanted %v ; got %v", w, newWav)
			continue
		}

		t.Logf("OK on index %d: %v", idx, w.Header)
	}
}

func TestWavOutputCompare(t *testing.T) {
	testdata, err := load()
	if err != nil {
		t.Error(err)
		return
	}

	for idx, test := range testdata {
		w := new(wav.Wav)
		_, err := w.Write(test.data)
		if err != nil {
			t.Error(err)
			return
		}
		b := w.Bytes()
		cmp := bytes.Compare(test.data, b)
		if cmp != 0 {
			t.Errorf("idx #%d: comparison failed with %d: len wants %d ; len got %d", idx, cmp, len(test.data), len(b))
			return
		}
	}
}

func TestWavWriteRead(t *testing.T) {
	testdata, err := load()
	if err != nil {
		t.Error(err)
		return
	}

	for idx, test := range testdata {
		w := new(wav.Wav)
		_, err := w.Write(test.data)
		if err != nil {
			t.Errorf("decoding error on index %d: %v", idx, err)
			continue
		}

		buf := make([]byte, len(test.data))
		_, err = w.Read(buf)
		if err != nil {
			t.Errorf("encoding error on index %d: %v", idx, err)
			continue
		}
		if len(buf) != len(test.data) {
			t.Errorf("output length mismatch error: wanted %d ; got %d", len(test.data), len(buf))
		}
		for i := 0; i < len(test.data); i++ {
			if buf[i] != test.data[i] {
				t.Errorf("byte mismatch on index %d: #%d wanted %d; got %d -- total len: %d", idx, i, test.data[i], buf[i], len(buf))
				return
			}
		}

		newWav := new(wav.Wav)
		_, err = newWav.Write(buf)
		if err != nil {
			t.Errorf("2nd-pass decoding error on index %d: %v", idx, err)
			continue
		}

		newBuf := make([]byte, len(test.data))

		_, err = newWav.Read(newBuf)
		if err != nil {
			t.Errorf("2nd-pass encoding error on index %d: %v", idx, err)
			continue
		}

		cmp := bytes.Compare(buf, newBuf)
		if cmp != 0 {
			t.Errorf("2nd-pass encoding mismatches the 1st-pass encoding: compare 1st-2nd: %d", cmp)
			continue
		}
		if !reflect.DeepEqual(w, newWav) {
			t.Errorf("output object mismatch error: wanted %v ; got %v", w, newWav)
			continue
		}

		t.Logf("OK on index %d: %v", idx, w.Header)
	}
}

func TestWavSegmentedWrite(t *testing.T) {
	testdata, err := load()
	if err != nil {
		t.Error(err)
		return
	}

	for idx, test := range testdata {
		w := new(wav.Wav)

		_, err := w.Write(test.data[:36])
		if err != nil {
			t.Errorf("decoding error on index %d: %v", idx, err)
			continue
		}

		// get the junk chunk and a portion of the data chunk
		_, err = w.Write(test.data[36:128])
		if err != nil {
			t.Errorf("decoding error on index %d: %v", idx, err)
			continue
		}

		// get the rest of the data
		_, err = w.Write(test.data[128:])
		if err != nil {
			t.Errorf("decoding error on index %d: %v", idx, err)
			continue
		}

		buf := make([]byte, len(test.data))
		_, err = w.Read(buf)
		if err != nil {
			t.Errorf("encoding error on index %d: %v", idx, err)
			continue
		}

		for i := range test.data {
			if test.data[i] != buf[i] {
				t.Errorf("encoding mismatches the original data on index #%d: wanted %v ; got %v", i, test.data[i], buf[i])
				continue
			}
		}

		t.Logf("OK on index %d: %v", idx, w.Header)
	}
}

func TestWav_WriteProcessRead(t *testing.T) {
	td, err := load()
	require.NoError(t, err)

	for idx, test := range td {
		// Write
		w := new(wav.Wav)
		_, err = w.Write(test.data)
		require.NoError(t, err)

		// Process
		w.Data.Apply(
			filters.PhaseFlip(),
			filters.PhaseFlip(),
		)

		// Read
		buf := make([]byte, len(test.data))
		_, err = w.Read(buf)
		require.NoError(t, err)
		require.Equal(t, buf, test.data)

		// Write read bytes
		newWav := new(wav.Wav)
		_, err = newWav.Write(buf)
		require.NoError(t, err)

		// Read and compare
		newBuf := make([]byte, len(test.data))
		_, err = newWav.Read(newBuf)
		require.NoError(t, err)
		require.Equal(t, buf, newBuf)
		require.Equal(t, w, newWav)

		t.Logf("OK on index %d: %v", idx, w.Header)
	}
}

func BenchmarkWav_WriteProcessRead(b *testing.B) {
	testdata, err := load()
	require.NoError(b, err)

	for _, test := range testdata[:4] {
		w := new(wav.Wav)
		buf := make([]byte, len(test.data))

		b.Run("All", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Write
				_, err = w.Write(test.data)
				if err != nil {
					b.Error(err)
					return
				}

				// Process
				w.Data.Apply(
					filters.PhaseFlip(),
					filters.PhaseFlip(),
				)

				// Read
				_, err = w.Read(buf)
				if err != nil {
					b.Error(err)
					return
				}
			}
		})

		b.Run("WriteRead", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Write
				_, err = w.Write(test.data)
				if err != nil {
					b.Error(err)
					return
				}

				// Read
				_, err = w.Read(buf)
				if err != nil {
					b.Error(err)
					return
				}
			}
		})

		b.Run("Write", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Write
				_, err = w.Write(test.data)
				if err != nil {
					b.Error(err)
					return
				}
			}
		})

		b.Run("Read", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				w = new(wav.Wav)
				// Write once
				_, err = w.Write(test.data)
				if err != nil {
					b.Error(err)
					return
				}
				b.StartTimer()

				// Read
				_, err = w.Read(buf)
				if err != nil {
					b.Error(err)
					return
				}
			}
		})
	}
}

func TestStream(t *testing.T) {
	td, err := load()
	require.NoError(t, err)

	t.Run("WriteAndRead", func(t *testing.T) {
		for idx, test := range td {
			var size = 64
			var cfg = &wav.StreamConfig{
				Size: wav.SizeConfig{
					Size: size,
				},
			}

			// Write
			w := wav.NewStream(cfg, func([]float64) error {
				return nil
			})

			_, err = w.Write(test.data)
			require.NoError(t, err, "index", idx)

			headerBytes := int(header.Size + dataheader.Size + w.Chunks[0].Header().Subchunk2Size + dataheader.Size)

			// 24bit will have a different size
			if w.Size != size {
				size = w.Size
			}

			buf := make([]byte, headerBytes+size)

			_, err = w.Read(buf)
			require.NoError(t, err, "index", idx)
			require.Equal(t, test.data[len(test.data)-size:], buf[headerBytes:], "index", idx)
		}
	})

	t.Run("ReadFromAndRead", func(t *testing.T) {
		for idx, test := range td {
			var size = 64
			var cfg = &wav.StreamConfig{
				Size: wav.SizeConfig{
					Size: size,
				},
			}

			// Write
			w := wav.NewStream(cfg, func([]float64) error {
				return nil
			})

			r := bytes.NewReader(test.data)
			_, err = w.ReadFrom(r)

			require.NoError(t, err, "index", idx)

			headerBytes := int(header.Size + dataheader.Size + w.Chunks[0].Header().Subchunk2Size + dataheader.Size)

			// 24bit will have a different size
			if w.Size != size {
				size = w.Size
			}

			buf := make([]byte, headerBytes+size)

			_, err = w.Read(buf)
			require.NoError(t, err, "index", idx)

			// compare last `size` bytes of the PCM data from input and data buffer
			require.Equal(t, test.data[len(test.data)-size:], buf[headerBytes:], "index", idx, "size", size)
		}
	})
}
