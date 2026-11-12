package audio

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync"
)

// DecodeFunc opens a decoder over r, which starts at the beginning of the stream.
type DecodeFunc func(r io.Reader) (SampleReader, error)

// ErrUnknownFormat is returned by Decode when no registered format recognises the input.
var ErrUnknownFormat = errors.New("audio: unknown format")

type registered struct {
	name   string
	magic  string
	decode DecodeFunc
}

//nolint:gochecknoglobals // the registry is process-wide by design, like image.RegisterFormat
var (
	formatsMu sync.RWMutex
	formats   []registered
)

// RegisterFormat registers a format for Decode. magic is the byte prefix that identifies it; a '?' matches any byte
// (e.g. "RIFF????WAVE"). It is usually called from a codec package's init function.
func RegisterFormat(name, magic string, decode DecodeFunc) {
	formatsMu.Lock()
	defer formatsMu.Unlock()

	formats = append(formats, registered{name: name, magic: magic, decode: decode})
}

// peeker is satisfied by *bufio.Reader.
type peeker interface {
	io.Reader
	Peek(n int) ([]byte, error)
}

// Decode identifies the format of r by its leading bytes and returns a decoder for it, with the format's registered
// name. A codec package must be imported (possibly for side effects only, e.g. `_ ".../encoding/wav"`) for its format
// to be recognised. Input too short or unlike any registered format returns ErrUnknownFormat; a read error is returned
// as is.
func Decode(r io.Reader) (SampleReader, string, error) {
	if r == nil {
		return nil, "", fmt.Errorf("%w: nil reader", ErrUnknownFormat)
	}

	p, ok := r.(peeker)
	if !ok {
		p = bufio.NewReader(r)
	}

	formatsMu.RLock()

	candidates := append([]registered(nil), formats...)

	formatsMu.RUnlock()

	// peek once, for the longest magic: a short input is just unknown, but a read error is reported as such
	size := 1
	for _, f := range candidates {
		size = max(size, len(f.magic))
	}

	head, err := p.Peek(size)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, bufio.ErrBufferFull) {
		return nil, "", err
	}

	for _, f := range candidates {
		if !match(f.magic, head) {
			continue
		}

		sr, err := f.decode(p)

		return sr, f.name, err
	}

	return nil, "", ErrUnknownFormat
}

// match reports whether head starts with magic, where '?' in magic matches any byte.
func match(magic string, head []byte) bool {
	if len(head) < len(magic) {
		return false
	}

	for i := range len(magic) {
		if magic[i] != '?' && magic[i] != head[i] {
			return false
		}
	}

	return true
}
