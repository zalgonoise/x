package randomizer

import (
	"crypto/rand"
)

type CryptoRand struct {
	size int
}

func (r CryptoRand) Random() ([]byte, error) {
	buf := make([]byte, r.size)

	if _, err := rand.Reader.Read(buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func New(size int) CryptoRand {
	return CryptoRand{
		size: size,
	}
}
