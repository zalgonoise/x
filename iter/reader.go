package iter

import (
	"errors"
	"io"
)

const minSize = 8

func ReadChunks(r io.Reader, size int) Seq[[]byte, error] {
	if size < minSize {
		size = minSize
	}

	return func(yield func([]byte, error) bool) bool {
		for {
			buf := make([]byte, size)
			n, err := r.Read(buf)

			if errors.Is(err, io.EOF) {
				if closer, ok := r.(io.Closer); ok {
					return closer.Close() == nil
				}

				return true
			}

			if err != nil {
				yield(nil, err)

				return false
			}

			if n == 0 {
				return true
			}

			if !yield(buf[:n], nil) {
				return false
			}
		}
	}
}
