package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"
)

// Cryptographer describes the set of cryptographic actions required by the app
type Cryptographer interface {
	// NewSalt generates a new random salt value, of 128 bytes in size
	NewSalt() [128]byte
	// NewKey generates a new random key value, of 256 bytes in size
	NewKey() [256]byte
	// NewCipher generates a new AES cipher based on the input key
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

// NewSalt generates a new random salt value, of 128 bytes in size
func (g *cryptographer) NewSalt() [128]byte {
	salt := [128]byte{}
	g.Lock()
	_, _ = g.random.Read(salt[:])
	g.Unlock()
	return salt
}

// NewKey generates a new random key value, of 256 bytes in size
func (g *cryptographer) NewKey() [256]byte {
	key := [256]byte{}
	g.Lock()
	_, _ = g.random.Read(key[:])
	g.Unlock()
	return key
}

// NewCipher generates a new AES cipher based on the input key
func (g *cryptographer) NewCipher(key []byte) (cipher.Block, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// NewSalt generates a new random salt value, of 128 bytes in size
func NewSalt() [128]byte {
	return crypt.NewSalt()
}

// NewKey generates a new random key value, of 256 bytes in size
func NewKey() [256]byte {
	return crypt.NewKey()
}

// NewCipher generates a new AES cipher based on the input key
func NewCipher(key []byte) (cipher.Block, error) {
	return crypt.NewCipher(key)
}
