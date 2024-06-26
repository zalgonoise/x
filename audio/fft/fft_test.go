package fft_test

import (
	"cmp"
	"math"
	"slices"
	"testing"
	"time"

	dspfft "github.com/mjibson/go-dsp/fft"
	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/osc"
)

// BenchmarkHypotenuse compares three approaches to calculating the hypotenuse using Go's standard library,
// to prioritize the most performant approach -- that is, by declaring variables for the real and imaginary parts of the
// complex number in question, and calculating `sqrt(real*real + imag*imag)`
//
//	❯ go test -bench . -benchmem -benchtime=5s -cpuprofile /tmp/cpu.pprof ./fft
//
//	goos: linux
//	goarch: amd64
//	pkg: github.com/zalgonoise/x/audio/wav/fft
//	cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
//	BenchmarkHypotenuse/Simplified-4                60037294                18.63 ns/op            0 B/op          0 allocs/op
//	BenchmarkHypotenuse/Minimal-4                   415254681                3.015 ns/op           0 B/op          0 allocs/op
//	BenchmarkHypotenuse/Extended-4                  330081352                3.616 ns/op           0 B/op          0 allocs/op
func BenchmarkHypotenuse(b *testing.B) {
	complexData := []complex128{
		0.5 + 1.3i, 1.1 + 0.4i, 0.8 - 1.2i, 1.3 - 0.5i,
	}

	b.Run("Simplified", func(b *testing.B) {
		var out [4]float64

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for idx := range complexData {
				out[idx] = math.Hypot(real(complexData[idx]), imag(complexData[idx]))
			}
		}

		_ = out
	})

	b.Run("Minimal", func(b *testing.B) {
		var out [4]float64

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for idx := range complexData {
				re := real(complexData[idx])
				im := imag(complexData[idx])
				out[idx] = math.Sqrt(re*re + im*im)
			}
		}

		_ = out
	})

	b.Run("Extended", func(b *testing.B) {
		var out [4]float64

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for idx := range complexData {
				out[idx] = math.Sqrt(
					real(complexData[idx])*real(complexData[idx]) +
						imag(complexData[idx])*imag(complexData[idx]),
				)
			}
		}

		_ = out
	})
}

func newSine(freq int) (*wav.Wav, error) {
	// create a sine wave 16 bit depth, 44.1kHz rate, mono,
	// 5 second duration. Pass audio bytes into a new bytes.Buffer
	sine, err := wav.New(44100, 16, 1, 1)
	if err != nil {
		return nil, err
	}

	sine.Generate(osc.SineWave, freq, 5*time.Second)

	return sine, nil
}

// BenchmarkFFT ensures that this library's FFT implementation yields the same results
// as go-dsp/fft, while running a comparison benchmark test to measure both implementations'
// performance
//
//	❯ go test -bench '^(BenchmarkFFT)$' -run='^$'  -benchmem -benchtime=5s -cpuprofile /tmp/cpu.pprof ./wav/fft
//	goos: linux
//	goarch: amd64
//	pkg: github.com/zalgonoise/x/audio/wav/fft
//	cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
//	BenchmarkFFT/Self/FFT-4                  4871833              1312 ns/op            1024 B/op          2 allocs/op
//	BenchmarkFFT/GoDSP/FFT-4                  213699             25352 ns/op            1803 B/op         26 allocs/op
//	BenchmarkFFT/Compare-4                  1000000000               0.0000034 ns/op               0 B/op          0 allocs/op
func BenchmarkFFT(b *testing.B) {
	sine, err := newSine(2000)
	if err != nil {
		b.Error(err)
		return
	}

	var (
		data      = fft.ToComplex(sine.Data.Float())[:32]
		spectrumA []complex128
		spectrumB []complex128
	)

	b.Run("Self/FFT", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			spectrumA = fft.FFT(data)
		}
		b.StopTimer()

		_ = spectrumA
	})

	b.Run("GoDSP/FFT", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			spectrumB = dspfft.FFT(data)
		}
		b.StopTimer()

		_ = spectrumB
	})

	b.Run("Compare", func(b *testing.B) {
		if len(spectrumA) != len(spectrumB) {
			b.Errorf("output length mismatch error: slice A: %d ; slice B: %d", len(spectrumA), len(spectrumB))
			return
		}

		for idx := range spectrumA {
			if spectrumA[idx] != spectrumB[idx] {
				b.Errorf("output mismatch error: index #%d: slice A: %v ; slice B: %v", idx, spectrumA[idx], spectrumB[idx])
			}
		}
	})
}

func TestApply(t *testing.T) {
	sine, err := newSine(2000)
	if err != nil {
		t.Error(err)
		return
	}

	data := sine.Data.Float()
	blockSize := 2048
	sampleRate := 44100
	windowBlock := window.Blackman2048
	wantsFreq := 2002

	spectrum := fft.Apply(
		sampleRate,
		data[:blockSize],
		windowBlock,
	)

	slices.SortFunc(spectrum, func(a, b fft.FrequencyPower) int {
		return cmp.Compare(b.Mag, a.Mag)
	})

	require.Equal(t, wantsFreq, spectrum[0].Freq)

	for i := range spectrum[:5] {
		t.Logf("freq: %v ; mag: %v", spectrum[i].Freq, spectrum[i].Mag)
	}
}

func TestConvolution(t *testing.T) {
	sine, err := newSine(2000)
	if err != nil {
		t.Error(err)
		return
	}

	blockSize := 16
	blockX := sine.Data.Float()[blockSize : blockSize*2]
	blockY := sine.Data.Float()[blockSize*2 : blockSize*3]

	convolution := fft.Convolve(fft.ToComplex(blockX), fft.ToComplex(blockY))

	t.Log(convolution)
}

func BenchmarkApply(b *testing.B) {
	sine, err := newSine(2000)
	if err != nil {
		b.Error(err)
		return
	}

	data := sine.Data.Float()
	blockSize := 256
	sampleRate := 44100
	windowBlock := window.Blackman256

	b.Run("2kSineFirstChunk", func(b *testing.B) {
		var spectrum []fft.FrequencyPower

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			spectrum = fft.Apply(
				sampleRate,
				data[:blockSize],
				windowBlock,
			)
		}

		_ = spectrum
	})
}
