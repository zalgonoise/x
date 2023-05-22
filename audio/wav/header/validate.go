package header

import (
	"fmt"

	"github.com/zalgonoise/x/audio/validation"
)

type SampleRate uint32

const (
	SampleRate44100  SampleRate = 44100
	SampleRate48000  SampleRate = 48000
	SampleRate88200  SampleRate = 88200
	SampleRate96000  SampleRate = 96000
	SampleRate176400 SampleRate = 176400
	SampleRate192000 SampleRate = 192000
)

type Channels uint16

const (
	ChannelsMono Channels = iota + 1
	ChannelsStereo
)

type BitDepth uint16

const (
	BitDepth8 BitDepth = 8 * (iota + 1)
	BitDepth16
	BitDepth24
	BitDepth32
)

type AudioFormat uint16

const (
	UnsetFormat AudioFormat = iota
	PCMFormat
	_
	FloatFormat
)

var (
	sampleRateValidator = validation.New[SampleRate](
		ErrInvalidSampleRate,
		SampleRate44100,
		SampleRate48000,
		SampleRate88200,
		SampleRate96000,
		SampleRate176400,
		SampleRate192000,
	)

	bitDepthValidator = validation.New[BitDepth](
		ErrInvalidBitDepth,
		BitDepth8,
		BitDepth16,
		BitDepth24,
		BitDepth32,
	)

	channelsValidator = validation.New[Channels](
		ErrInvalidNumChannels,
		ChannelsMono,
		ChannelsStereo,
	)

	audioFormatValidator = validation.New[AudioFormat](
		ErrInvalidAudioFormat,
		PCMFormat,   // PCM audio
		FloatFormat, // IEEE floating-point 32-bit audio
	)
)

func Validate(header *Header) error {
	if header == nil {
		return ErrEmptyHeader
	}

	if string(header.ChunkID[:]) != string(defaultChunkID[:]) {
		return fmt.Errorf("%w: ChunkID %s", ErrInvalidHeader, string(header.ChunkID[:]))
	}

	if string(header.Format[:]) != string(defaultFormat[:]) {
		return fmt.Errorf("%w: Format %s", ErrInvalidHeader, string(header.Format[:]))
	}

	if err := sampleRateValidator.Validate(SampleRate(header.SampleRate)); err != nil {
		return fmt.Errorf("%w: SampleRate %d", err, header.SampleRate)
	}

	if err := bitDepthValidator.Validate(BitDepth(header.BitsPerSample)); err != nil {
		return fmt.Errorf("%w: BitsPerSample %d", err, header.BitsPerSample)
	}

	if err := channelsValidator.Validate(Channels(header.NumChannels)); err != nil {
		return fmt.Errorf("%w: NumChannels %d", err, header.NumChannels)
	}

	if err := audioFormatValidator.Validate(AudioFormat(header.AudioFormat)); err != nil {
		return fmt.Errorf("%w: AudioFormat %d", err, header.AudioFormat)
	}

	return nil
}
