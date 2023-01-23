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
	// New256Key generates a new random key value, of 256 bytes in size
	New256Key() [256]byte
	// New32Key generates a new random key value, of 256 bytes in size
	New32Key() [32]byte
	// NewCipher generates a new AES cipher based on the input key
	NewCipher(key []byte) EncryptDecrypter
	// Random reads random bytes into the input byte slice
	Random(buffer []byte)
}

type EncryptDecrypter interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type aesEncrypter struct {
	key []byte
}

func init() {
	cryptog = &cryptographer{}
	var rngSeed int64
	_ = binary.Read(cryptorand.Reader, binary.LittleEndian, &rngSeed)
	cryptog.(*cryptographer).random = rand.New(rand.NewSource(rngSeed))
}

var cryptog Cryptographer

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

// New256Key generates a new random key value, of 256 bytes in size
func (g *cryptographer) New256Key() [256]byte {
	key := [256]byte{}
	g.Lock()
	_, _ = g.random.Read(key[:])
	g.Unlock()
	return key
}

// New32Key generates a new random key value, of 256 bytes in size
func (g *cryptographer) New32Key() [32]byte {
	key := [32]byte{}
	g.Lock()
	_, _ = g.random.Read(key[:])
	g.Unlock()
	return key
}

func (g *cryptographer) Random(buffer []byte) {
	g.Lock()
	_, _ = g.random.Read(buffer)
	g.Unlock()
}

// NewCipher generates a new AES cipher based on the input key
func (g *cryptographer) NewCipher(key []byte) EncryptDecrypter {
	return aesEncrypter{
		key: key,
	}
}

func (enc aesEncrypter) Encrypt(plaintext []byte) ([]byte, error) {
	c, err := aes.NewCipher(enc.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	cryptog.Random(nonce)

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (enc aesEncrypter) Decrypt(ciphertext []byte) ([]byte, error) {
	c, err := aes.NewCipher(enc.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// NewSalt generates a new random salt value, of 128 bytes in size
func NewSalt() [128]byte {
	return cryptog.NewSalt()
}

// NewKey generates a new random key value, of 256 bytes in size
func New256Key() [256]byte {
	return cryptog.New256Key()
}

// NewKey generates a new random key value, of 256 bytes in size
func New32Key() [32]byte {
	return cryptog.New32Key()
}

// NewCipher generates a new AES cipher based on the input key
func NewCipher(key []byte) EncryptDecrypter {
	return cryptog.NewCipher(key)
}
