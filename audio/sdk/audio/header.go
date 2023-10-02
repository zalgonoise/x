package audio

// Header describes types considered headers on different audio encodings.
//
// Methods implemented by a Header return data that is common between all
// audio encoding headers. This interface type is used in order to ensure that different
// audio encodings can be processed similarly when it comes to extracting and processing
// audio data.
type Header interface {
	// GetSampleRate returns the sample rate for this audio signal.
	//
	// An unset (and also invalid) value is zero. This should be the value returned when the
	// Header is nil, when the Header hasn't yet been read or parsed, or when the Header
	// implementation is a no-op.
	GetSampleRate() int
}
