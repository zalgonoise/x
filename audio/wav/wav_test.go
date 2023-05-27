package wav_test

import (
	"bytes"
	"embed"
	_ "embed"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/header"
)

//go:embed testdata/*
var testdataFS embed.FS

func load() ([][]byte, error) {
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

	if mono8bit44100, err = testdataFS.ReadFile("testdata/amen_kick_mono_8bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if mono16bit44100, err = testdataFS.ReadFile("testdata/amen_kick_mono_16bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if mono24bit44100, err = testdataFS.ReadFile("testdata/amen_kick_mono_24bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if mono32bit44100, err = testdataFS.ReadFile("testdata/amen_kick_mono_32bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if mono32bit96000, err = testdataFS.ReadFile("testdata/amen_kick_mono_32bit_96000hz.wav"); err != nil {
		return nil, err
	}
	if mono32bit192000, err = testdataFS.ReadFile("testdata/amen_kick_mono_32bit_192000hz.wav"); err != nil {
		return nil, err
	}
	if mono8bit176400, err = testdataFS.ReadFile("testdata/amen_kick_mono_8bit_176400hz.wav"); err != nil {
		return nil, err
	}
	if stereo8bit44100, err = testdataFS.ReadFile("testdata/amen_kick_stereo_8bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if stereo16bit44100, err = testdataFS.ReadFile("testdata/amen_kick_stereo_16bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if stereo24bit44100, err = testdataFS.ReadFile("testdata/amen_kick_stereo_24bit_44100hz.wav"); err != nil {
		return nil, err
	}
	if stereo32bit44100, err = testdataFS.ReadFile("testdata/amen_kick_stereo_32bit_44100hz.wav"); err != nil {
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

	return [][]byte{
		mono8bit44100,
		mono16bit44100,
		mono24bit44100,
		mono32bit44100,
		mono32bit96000,
		mono32bit192000,
		mono8bit176400, // remove a useless nullbyte in the end
		stereo8bit44100,
		stereo16bit44100,
		stereo24bit44100,
		stereo32bit44100,
	}, nil
}

func BenchmarkWav(b *testing.B) {
	td, err := load()
	if err != nil {
		b.Error(err)
		return
	}

	for _, testdata := range td[:4] { // mono 44.1kHz 8bit to 32bit
		var loadedWav *wav.Wav

		loadedWav, err = wav.Decode(testdata)
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
					w, err = wav.Decode(testdata)
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
					_, err = w.Write(testdata)
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
					buf := make([]byte, len(testdata))
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
		w, err := wav.Decode(test)
		if err != nil {
			t.Errorf("decoding error on index %d: %v", idx, err)
			continue
		}
		buf := w.Bytes()
		if len(buf) != len(test) {
			t.Errorf("output length mismatch error on index %d: wanted %d ; got %d", idx, len(test), len(buf))

		}
		for i := 0; i < len(buf); i++ {
			if buf[i] != test[i] {
				t.Errorf("byte mismatch on index %d: #%d wanted %d; got %d -- total len: %d", idx, i, test[i], buf[i], len(buf))
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
		_, err := w.Write(test)
		if err != nil {
			t.Error(err)
			return
		}
		b := w.Bytes()
		cmp := bytes.Compare(test, b)
		if cmp != 0 {
			t.Errorf("idx #%d: comparison failed with %d: len wants %d ; len got %d", idx, cmp, len(test), len(b))
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
		_, err := w.Write(test)
		if err != nil {
			t.Errorf("decoding error on index %d: %v", idx, err)
			continue
		}

		buf := make([]byte, len(test))
		_, err = w.Read(buf)
		if err != nil {
			t.Errorf("encoding error on index %d: %v", idx, err)
			continue
		}
		if len(buf) != len(test) {
			t.Errorf("output length mismatch error: wanted %d ; got %d", len(test), len(buf))
		}
		for i := 0; i < len(test); i++ {
			if buf[i] != test[i] {
				t.Errorf("byte mismatch on index %d: #%d wanted %d; got %d -- total len: %d", idx, i, test[i], buf[i], len(buf))
				return
			}
		}

		newWav := new(wav.Wav)
		_, err = newWav.Write(buf)
		if err != nil {
			t.Errorf("2nd-pass decoding error on index %d: %v", idx, err)
			continue
		}

		newBuf := make([]byte, len(test))

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

		_, err := w.Write(test[:36])
		if err != nil {
			t.Errorf("decoding error on index %d: %v", idx, err)
			continue
		}

		// get the junk chunk and a portion of the data chunk
		_, err = w.Write(test[36:128])
		if err != nil {
			t.Errorf("decoding error on index %d: %v", idx, err)
			continue
		}

		// get the rest of the data
		_, err = w.Write(test[128:])
		if err != nil {
			t.Errorf("decoding error on index %d: %v", idx, err)
			continue
		}

		buf := make([]byte, len(test))
		_, err = w.Read(buf)
		if err != nil {
			t.Errorf("encoding error on index %d: %v", idx, err)
			continue
		}

		for i := range test {
			if test[i] != buf[i] {
				t.Errorf("encoding mismatches the original data on index #%d: wanted %v ; got %v", i, test[i], buf[i])
				continue
			}
		}

		t.Logf("OK on index %d: %v", idx, w.Header)
	}

}
