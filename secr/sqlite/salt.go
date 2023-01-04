package sqlite

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"
)

type SaltGenerator interface {
	NewSalt() [128]byte
}

func init() {
	saltGen = &saltGenerator{}
	var rngSeed int64
	_ = binary.Read(cryptorand.Reader, binary.LittleEndian, &rngSeed)
	saltGen.(*saltGenerator).random = rand.New(rand.NewSource(rngSeed))
}

var saltGen SaltGenerator

type saltGenerator struct {
	sync.Mutex
	random *rand.Rand
}

func (g *saltGenerator) NewSalt() [128]byte {
	salt := [128]byte{}
	g.Lock()
	_, _ = g.random.Read(salt[:])
	g.Unlock()
	return salt
}

func NewSalt() [128]byte {
	return saltGen.NewSalt()
}
