package audio

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"testing"

	"github.com/go-audio/audio"
	gwav "github.com/go-audio/wav"
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

type testData struct {
	name        string
	data        []byte
	sampleRate  int
	bitDepth    int
	numChannels int
}

var testdata = []testData{
	{
		name:        "Mono8Bit44100Hz",
		data:        mono8bit44100,
		sampleRate:  44100,
		bitDepth:    8,
		numChannels: 1,
	},
	{
		name:        "Mono16Bit44100Hz",
		data:        mono16bit44100,
		sampleRate:  44100,
		bitDepth:    16,
		numChannels: 1,
	},
	{
		name:        "Mono24Bit44100Hz",
		data:        mono24bit44100,
		sampleRate:  44100,
		bitDepth:    24,
		numChannels: 1,
	},
	{
		name:        "Mono32Bit44100Hz",
		data:        mono32bit44100,
		sampleRate:  44100,
		bitDepth:    32,
		numChannels: 1,
	},
	{
		name:        "Mono32Bit96000Hz",
		data:        mono32bit96000,
		sampleRate:  96000,
		bitDepth:    32,
		numChannels: 1,
	},
	{
		name:        "Mono32Bit192000Hz",
		data:        mono32bit192000,
		sampleRate:  192000,
		bitDepth:    32,
		numChannels: 1,
	},
	{
		name:        "Mono8Bit176400Hz",
		data:        mono8bit176400,
		sampleRate:  176400,
		bitDepth:    8,
		numChannels: 1,
	},
	{
		name:        "Stereo8Bit44100Hz",
		data:        stereo8bit44100,
		sampleRate:  44100,
		bitDepth:    8,
		numChannels: 2,
	},
	{
		name:        "Stereo16Bit44100Hz",
		data:        stereo16bit44100,
		sampleRate:  44100,
		bitDepth:    16,
		numChannels: 2,
	},
	{
		name:        "Stereo24Bit44100Hz",
		data:        stereo24bit44100,
		sampleRate:  44100,
		bitDepth:    24,
		numChannels: 2,
	},
	{
		name:        "Stereo32Bit44100Hz",
		data:        stereo32bit44100,
		sampleRate:  44100,
		bitDepth:    32,
		numChannels: 2,
	},
}

func BenchmarkGoAudioWav(b *testing.B) {
	for _, d := range testdata {
		b.Run(
			fmt.Sprintf("Decode%s", d.name), func(b *testing.B) {
				var (
					pcm *audio.IntBuffer
					err error
				)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					r := bytes.NewReader(d.data)
					dec := gwav.NewDecoder(r)
					pcm, err = dec.FullPCMBuffer()
					if err != nil {
						b.Error(err)
						return
					}
				}
				_ = pcm
			},
		)

		b.Run(
			fmt.Sprintf("Encode%s", d.name), func(b *testing.B) {
				var (
					pcm *audio.IntBuffer
					err error
				)
				r := bytes.NewReader(d.data)
				dec := gwav.NewDecoder(r)
				pcm, err = dec.FullPCMBuffer()
				if err != nil {
					b.Error(err)
					return
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					tmp, err := os.CreateTemp("/tmp/", "enc")
					if err != nil {
						b.Error(err)
						return
					}
					enc := gwav.NewEncoder(tmp, d.sampleRate, d.bitDepth, d.numChannels, 1)
					err = enc.Write(pcm)
					if err != nil {
						b.Error(err)
						return
					}
					_ = os.RemoveAll(tmp.Name())
				}
			},
		)
	}
}

func BenchmarkXAudioWav(b *testing.B) {
	for _, d := range testdata {
		b.Run(
			fmt.Sprintf("Decode%s", d.name), func(b *testing.B) {
				var (
					w   *Wav
					err error
				)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					w, err = Decode(d.data)
					if err != nil {
						b.Error(err)
					}
				}
				_ = w
			},
		)
		b.Run(
			fmt.Sprintf("Encode%s", d.name), func(b *testing.B) {
				var (
					w   *Wav
					err error
					buf []byte
				)

				w, err = Decode(d.data)
				if err != nil {
					b.Error(err)
					return
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					buf = w.Bytes()
				}
				_ = buf
			},
		)
	}
}
