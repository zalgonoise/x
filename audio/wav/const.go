package wav

const (
	junkSubchunkIDString = "junk"

	headerLen = 36
)

var (
	defaultChunkID     = [4]byte{82, 73, 70, 70}
	defaultFormat      = [4]byte{87, 65, 86, 69}
	defaultSubchunk1ID = [4]byte{102, 109, 116, 32}
)

const (
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

var (
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
		1: {}, // PCM audio
		//3: {}, // IEEE floating-point 32-bit audio
	}
)
