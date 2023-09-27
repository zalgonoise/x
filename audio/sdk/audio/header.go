package audio

// Header describes types considered headers on different audio encodings.
//
// Methods implemented by a Header return data that is common between all
// audio encoding headers.
type Header interface {
	// GetSampleRate returns the sample rate for this audio signal.
	GetSampleRate() int
}

type noOpHeader struct{}

// GetSampleRate implements the Header interface
//
// This is a no-op call and the returned sample rate is always zero
func (noOpHeader) GetSampleRate() int { return 0 }

// NoOpHeader returns a no-op Header
func NoOpHeader() Header {
	return noOpHeader{}
}
