package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"
)

type Cryptographer interface {
	NewSalt() [128]byte
	NewKey() [256]byte
	NewCipher(key []byte) (cipher.Block, error)
}

func init() {
	crypt = &cryptographer{}
	var rngSeed int64
	_ = binary.Read(cryptorand.Reader, binary.LittleEndian, &rngSeed)
	crypt.(*cryptographer).random = rand.New(rand.NewSource(rngSeed))
}

var crypt Cryptographer

type cryptographer struct {
	sync.Mutex
	random *rand.Rand
}

func (g *cryptographer) NewSalt() [128]byte {
	salt := [128]byte{}
	g.Lock()
	_, _ = g.random.Read(salt[:])
	g.Unlock()
	return salt
}

func (g *cryptographer) NewKey() [256]byte {
	key := [256]byte{}
	g.Lock()
	_, _ = g.random.Read(key[:])
	g.Unlock()
	return key
}

func (g *cryptographer) NewCipher(key []byte) (cipher.Block, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func NewSalt() [128]byte {
	return crypt.NewSalt()
}

func NewKey() [256]byte {
	return crypt.NewKey()
}

func NewCipher(key []byte) (cipher.Block, error) {
	return crypt.NewCipher(key)
}
