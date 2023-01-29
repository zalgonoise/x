package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"math/rand"
	"sync"
	"time"

	"github.com/zalgonoise/x/errors"
	"golang.org/x/crypto/pbkdf2"
)

const numHashIter = 600_001

var (
	ErrInvalidLen = errors.New("invalid ciphertext length")
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

// EncrypterDecrypter is a general purpose encryption interface that supports
// an Encrypt and a Decrypt method
type EncryptDecrypter interface {
	// Encrypt will encrypt the input bytes `v` with the EncryptDecrypter key,
	// returning the ciphertext of `v` as a byte slice, and an error
	Encrypt(v []byte) ([]byte, error)

	// Decrypt will decipher the input bytes `v` with the EncryptDecrypter key,
	// returning the plaintext of `v` as a byte slice, and an error
	Decrypt(v []byte) ([]byte, error)
}

type cryptographer struct {
	sync.Mutex
	random *rand.Rand
}

var cryptog Cryptographer

func init() {
	cryptog = &cryptographer{}
	var rngSeed int64 = time.Now().Unix()
	_ = binary.Read(cryptorand.Reader, binary.LittleEndian, &rngSeed)
	cryptog.(*cryptographer).random = rand.New(rand.NewSource(rngSeed))
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

// Random reads random bytes into the input byte slice
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

type aesEncrypter struct {
	key []byte
}

// Encrypt will encrypt the input bytes `v` with the EncryptDecrypter key,
// returning the ciphertext of `v` as a byte slice, and an error
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

// Decrypt will decipher the input bytes `v` with the EncryptDecrypter key,
// returning the plaintext of `v` as a byte slice, and an error
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
		return nil, ErrInvalidLen
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

func Hash(secret, salt []byte) []byte {
	return pbkdf2.Key(secret, salt, numHashIter, 128, sha512.New)
}
