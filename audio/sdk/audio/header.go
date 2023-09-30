package audio

// Header describes types considered headers on different audio encodings.
//
// Methods implemented by a Header return data that is common between all
// audio encoding headers.
type Header interface {
	// GetSampleRate returns the sample rate for this audio signal.
	GetSampleRate() int
}
