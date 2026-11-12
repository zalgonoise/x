// Package iox holds io helpers shared by the codec packages.
package iox

import (
	"errors"
	"io"
)

// MaxEmptyReads is how many consecutive reads returning (0, nil) are tolerated before giving up with io.ErrNoProgress,
// as bufio does.
const MaxEmptyReads = 100

// ReadAtLeast reads from r into buf until it has read at least atLeast bytes, like io.ReadAtLeast, with two
// differences:
//
//   - The end of the input is always reported as io.EOF, even after a partial read (n > 0). io.ReadAtLeast reports that
//     case as io.ErrUnexpectedEOF, which makes it indistinguishable from an io.ErrUnexpectedEOF returned by r itself,
//     such as an HTTP body whose connection dropped. Here, every error other than io.EOF comes from r.
//   - A reader returning (0, nil) MaxEmptyReads times in a row makes it fail with io.ErrNoProgress.
//
// Once atLeast bytes are read, an io.EOF returned with them is dropped (the next read reports it again); any other
// error is returned with them. atLeast must not exceed len(buf).
func ReadAtLeast(r io.Reader, buf []byte, atLeast int) (n int, err error) {
	for empty := 0; n < atLeast; {
		var nn int

		nn, err = r.Read(buf[n:])
		n += nn

		switch {
		case err != nil:
			if n >= atLeast && errors.Is(err, io.EOF) {
				return n, nil
			}

			return n, err
		case nn > 0:
			empty = 0
		default:
			empty++

			if empty >= MaxEmptyReads {
				return n, io.ErrNoProgress
			}
		}
	}

	return n, nil
}

// Progress wraps r so that MaxEmptyReads consecutive reads returning (0, nil) fail with io.ErrNoProgress, which keeps
// io.Copy and similar loops from spinning forever on a misbehaving reader.
func Progress(r io.Reader) io.Reader {
	if p, ok := r.(*progress); ok {
		return p
	}

	return &progress{r: r}
}

type progress struct {
	r     io.Reader
	empty int
}

func (p *progress) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)

	switch {
	case n > 0 || err != nil || len(b) == 0:
		p.empty = 0
	default:
		p.empty++

		if p.empty >= MaxEmptyReads {
			p.empty = 0

			return 0, io.ErrNoProgress
		}
	}

	return n, err
}
