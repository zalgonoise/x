package header

import (
	"fmt"
)

const (
	headerLen = 36

	sampleRate44100  uint32 = 44100
	sampleRate48000  uint32 = 48000
	sampleRate88200  uint32 = 88200
	sampleRate96000  uint32 = 96000
	sampleRate176400 uint32 = 176400
	sampleRate192000 uint32 = 192000

	bitDepth8  uint16 = 8
	bitDepth16 uint16 = 16
	bitDepth24 uint16 = 24
	bitDepth32 uint16 = 32

	channelsMono   uint16 = 1
	channelsStereo uint16 = 2
)

type AudioFormat uint16

const (
	UnsetFormat AudioFormat = iota
	PCMFormat
	_
	FloatFormat
)

var (
	defaultChunkID = [4]byte{82, 73, 70, 70}
	defaultFormat  = [4]byte{87, 65, 86, 69}

	validSampleRates = map[uint32]struct{}{
		sampleRate44100:  {},
		sampleRate48000:  {},
		sampleRate88200:  {},
		sampleRate96000:  {},
		sampleRate176400: {},
		sampleRate192000: {},
	}

	validBitDepths = map[uint16]struct{}{
		bitDepth8:  {},
		bitDepth16: {},
		bitDepth24: {},
		bitDepth32: {},
	}

	validNumChannels = map[uint16]struct{}{
		channelsMono:   {},
		channelsStereo: {},
	}

	validAudioFormats = map[uint16]struct{}{
		uint16(PCMFormat):   {}, // PCM audio
		uint16(FloatFormat): {}, // IEEE floating-point 32-bit audio
	}
)

func Validate(header *Header) error {
	if string(header.ChunkID[:]) != string(defaultChunkID[:]) {
		return fmt.Errorf("%w: ChunkID %s", ErrInvalidHeader, string(header.ChunkID[:]))
	}

	if string(header.Format[:]) != string(defaultFormat[:]) {
		return fmt.Errorf("%w: Format %s", ErrInvalidHeader, string(header.Format[:]))
	}

	if _, ok := validSampleRates[header.SampleRate]; !ok {
		return fmt.Errorf("%w: SampleRate %d", ErrInvalidSampleRate, header.SampleRate)
	}

	if _, ok := validBitDepths[header.BitsPerSample]; !ok {
		return fmt.Errorf("%w: BitsPerSample %d", ErrInvalidBitDepth, header.BitsPerSample)
	}

	if _, ok := validNumChannels[header.NumChannels]; !ok {
		return fmt.Errorf("%w: NumChannels %d", ErrInvalidNumChannels, header.NumChannels)
	}

	if _, ok := validAudioFormats[header.AudioFormat]; !ok {
		return fmt.Errorf("%w: AudioFormat %d", ErrInvalidAudioFormat, header.AudioFormat)
	}

	return nil
}
