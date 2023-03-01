package wav

type err string

func (e err) Error() string { return (string)(e) }

const (
	ErrInvalidNumChannels err = "audio/wav: invalid number of channels"
	ErrInvalidSampleRate  err = "audio/wav: invalid sample rate"
	ErrInvalidBitDepth    err = "audio/wav: invalid bit depth"
	ErrInvalidHeader      err = "audio/wav: invalid WAV header"
	ErrShortDataBuffer    err = "audio/wav: data buffer is too short"
	ErrShortHeaderBuffer  err = "audio/wav: header buffer is too short"
	ErrZeroChunks         err = "audio/wav: no buffered chunks available"
	ErrMissingHeader      err = "audio/wav: missing header metadata"
	ErrMissingDataBuffer  err = "audio/wav: missing data buffer to write to, as no previous header was captured"
	ErrNegativeRead       err = "audio/wav: reader returned negative count from Read"
)
