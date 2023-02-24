package wav_test

import (
	"bytes"
	"reflect"
	"testing"

	_ "embed"

	. "github.com/zalgonoise/x/audio/wav"
)

//go:embed testdata/amen_kick_mono_8bit_44100hz.wav
var mono8bit44100 []byte

//go:embed testdata/amen_kick_mono_16bit_44100hz.wav
var mono16bit44100 []byte

//go:embed testdata/amen_kick_mono_24bit_44100hz.wav
var mono24bit44100 []byte

//go:embed testdata/amen_kick_mono_32bit_44100hz.wav
var mono32bit44100 []byte

//go:embed testdata/amen_kick_mono_32bit_96000hz.wav
var mono32bit96000 []byte

//go:embed testdata/amen_kick_mono_32bit_192000hz.wav
var mono32bit192000 []byte

//go:embed testdata/amen_kick_mono_8bit_176400hz.wav
var mono8bit176400 []byte

//go:embed testdata/amen_kick_stereo_8bit_44100hz.wav
var stereo8bit44100 []byte

//go:embed testdata/amen_kick_stereo_16bit_44100hz.wav
var stereo16bit44100 []byte

//go:embed testdata/amen_kick_stereo_24bit_44100hz.wav
var stereo24bit44100 []byte

//go:embed testdata/amen_kick_stereo_32bit_44100hz.wav
var stereo32bit44100 []byte

var testdata = [][]byte{
	mono8bit44100,
	mono16bit44100,
	mono24bit44100,
	mono32bit44100,
	mono32bit96000,
	mono32bit192000,
	mono8bit176400,
	stereo8bit44100,
	stereo16bit44100,
	stereo32bit44100,
}

func TestNewWav(t *testing.T) {
	wants := []byte{
		82, 73, 70, 70, 0, 0, 0, 0, 87, 65, 86, 69,
		102, 109, 116, 32, 16, 0, 0, 0, 1, 0, 2, 0,
		68, 172, 0, 0, 16, 177, 2, 0, 4, 0, 16, 0,
	}
	wav, err := New(44100, 16, 2)
	if err != nil {
		t.Error(err)
	}
	if string(wants) != string(wav.Header.Bytes()) {
		t.Errorf("output mismatch error: \n\nwanted %v ;\n\ngot %v\n", wants, wav.Header.Bytes())
	}

	parsedHeader, err := HeaderFrom(wav.Header.Bytes())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(parsedHeader, wav.Header) {
		t.Errorf("output mismatch error: \n\nwanted %v ;\n\ngot %v\n", wav.Header, parsedHeader)
	}
}

func TestFromFile(t *testing.T) {
	for idx, test := range testdata {
		wav, err := Decode(test)
		if err != nil {
			t.Errorf("decoding error on index %d: %v", idx, err)
			continue
		}
		buf, err := wav.Bytes()
		if err != nil {
			t.Errorf("encoding error on index %d: %v", idx, err)
			continue
		}
		if len(buf) != len(test) {
			t.Errorf("output length mismatch error: wanted %d ; got %d", len(test), len(buf))
		}
		for i := 0; i < len(test); i++ {
			if buf[i] != test[i] {
				t.Errorf("byte mismatch on index %d: wanted %d; got %d -- total len: %d", i, test[i], buf[i], len(buf))
				continue
			}
		}

		newWav, err := Decode(buf)
		if err != nil {
			t.Errorf("2nd-pass decoding error on index %d: %v", idx, err)
			continue
		}
		newBuf, err := newWav.Bytes()
		if err != nil {
			t.Errorf("2nd-pass encoding error on index %d: %v", idx, err)
			continue
		}

		cmp := bytes.Compare(buf, newBuf)
		if cmp != 0 {
			t.Errorf("2nd-pass encoding mismatches the 1st-pass encoding: compare 1st-2nd: %d", cmp)
			continue
		}
		if !reflect.DeepEqual(wav, newWav) {
			t.Errorf("output object mismatch error: wanted %v ; got %v", wav, newWav)
			continue
		}

		t.Logf("OK on index %d: %v", idx, wav.Header)
	}
}
