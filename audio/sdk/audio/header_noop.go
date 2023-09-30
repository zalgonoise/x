package audio

type noOpHeader struct{}

// GetSampleRate implements the Header interface
//
// This is a no-op call and the returned sample rate is always zero
func (noOpHeader) GetSampleRate() int { return 0 }

// NoOpHeader returns a no-op Header
func NoOpHeader() Header {
	return noOpHeader{}
}
