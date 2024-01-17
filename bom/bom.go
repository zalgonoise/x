package bom

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var ErrUnsupportedEncoding = errors.New("unsupported encoding")

// BOM describes the characteristics of a Byte Order Mark sequence of bytes, found in the
// head of (some) data and documents, which allows different applications to process it following
// the correct encoding and endianness.
//
// Described in: https://en.wikipedia.org/wiki/Byte_order_mark
type BOM struct {
	// Unicode represents whether the encoding is Unicode-compliant or not.
	Unicode bool
	// Size points to the integer size in each Unicode character.
	Size uint8
	// Header provides access to the BOM bytes used in this encoding
	Header []byte
	// ByteOrder points to the correct binary.ByteOrder if specified in the encoding.
	ByteOrder binary.ByteOrder
}

// Read scans up to the first four bytes of the input byte slice, to find a matching BOM header, in
// a read-only approach.
//
// The returned values are the number of bytes that make up the BOM Header if present, and an error
// is returned if the encoding is unsupported. A zero-value, nil-error return means that the byte
// slice does not contain a BOM Header.
func (b *BOM) Read(p []byte) (n int, err error) {
	if len(p) < 2 {
		return 0, nil
	}

	// similar headers, check UTF32 first
	if len(p) >= 4 {
		switch {
		case equal(bomUTF32BE, p[0:4]):
			b.Unicode = true
			b.Size = 32
			b.Header = bomUTF32BE
			b.ByteOrder = binary.BigEndian

			return 4, nil
		case equal(bomUTF32LE, p[0:4]):
			b.Unicode = true
			b.Size = 32
			b.Header = bomUTF32LE
			b.ByteOrder = binary.LittleEndian

			return 4, nil
		}
	}

	switch {
	case equal(bomUTF16BE, p[0:2]):
		b.Unicode = true
		b.Size = 16
		b.Header = bomUTF16BE
		b.ByteOrder = binary.BigEndian

		return 2, nil
	case equal(bomUTF16LE, p[0:2]):
		b.Unicode = true
		b.Size = 16
		b.Header = bomUTF16LE
		b.ByteOrder = binary.LittleEndian

		return 2, nil
	}

	if len(p) < 3 {
		return 0, nil
	}

	switch {
	case equal(bomUTF8, p[0:3]):
		b.Unicode = true
		b.Size = 8
		b.Header = bomUTF8

		return 3, nil
	case equal(bomUTF7, p[0:3]):
		b.Unicode = true
		b.Header = bomUTF7

		return 3, fmt.Errorf("%w: UTF-7", ErrUnsupportedEncoding)
	case equal(bomUTF1, p[0:3]):
		b.Unicode = true
		b.Header = bomUTF1

		return 3, fmt.Errorf("%w: UTF-1", ErrUnsupportedEncoding)
	case equal(bomSCSU, p[0:3]):
		b.Unicode = true
		b.Header = bomSCSU

		return 3, fmt.Errorf("%w: Standard Compression Scheme for Unicode", ErrUnsupportedEncoding)
	case equal(bomBOCU1, p[0:3]):
		b.Unicode = true
		b.Header = bomBOCU1

		return 3, fmt.Errorf("%w: Binary Ordered Compression for Unicode", ErrUnsupportedEncoding)
	}

	if len(p) < 4 {
		return 0, nil
	}

	switch {
	case equal(bomUTFEBCDIC, p[0:4]):
		b.Unicode = true
		b.Header = bomUTFEBCDIC

		return 4, fmt.Errorf("%w: UTF-EBCDIC", ErrUnsupportedEncoding)
	case equal(bomGB18030, p[0:4]):
		b.Unicode = true
		b.Header = bomGB18030

		return 4, fmt.Errorf("%w: GB 18030", ErrUnsupportedEncoding)
	}

	return 0, nil
}

// ReadFrom scans up to the first four bytes from the input io.Reader, to find a matching BOM header, in
// a read-only approach. This is done by leveraging a bufio.Reader's Peek method.
//
// The resulting byte slice from this call is then passed into this BOM's Read method.
//
// The returned values are the number of bytes that make up the BOM Header if present, and an error
// is returned if the encoding is unsupported. A zero-value, nil-error return means that the byte
// slice does not contain a BOM Header.
func (b *BOM) ReadFrom(r io.Reader) (n int64, err error) {
	buf := bufio.NewReader(r)

	head, err := buf.Peek(4)

	if err != nil {
		return 0, err
	}

	num, err := b.Read(head)

	return int64(num), err
}

// Drop removes the BOM from a byte slice, if present. This call creates a BOM instance, to Read the byte slice.
//
// If an error is raised, the byte slice is returned as-is without any changes; along with the number of BOM header
// bytes, and the error. Otherwise, the input byte slice is cropped from its head by the number of BOM header bytes;
// along with this value and a nil error.
func Drop(p []byte) ([]byte, int, error) {
	bom := &BOM{}

	n, err := bom.Read(p)

	if err != nil {
		return p, n, err
	}

	return p[n:], n, nil
}

// Discard is similar to Drop, but operates on an io.Reader type. This call creates a bufio.Reader to work with the
// input one as it is able to Peek into the first four bytes (without ever advancing the io.Reader's cursor).
//
// Then, with a new BOM instance, it calls its Read method on the peeked header, returning any error if raised. If no
// error is raised and the number of BOM header bytes is over zero; the bufio.Reader's Discard method.
//
// The returned io.Reader is the bufio.Reader that analyzes this byte stream; the number of BOM header bytes found; and
// an error if raised.
func Discard(r io.Reader) (io.Reader, int, error) {
	bom := &BOM{}
	buf := bufio.NewReader(r)

	head, err := buf.Peek(4)
	if err != nil {
		return buf, 0, err
	}

	n, err := bom.Read(head)

	if err != nil {
		return buf, n, err
	}

	if n > 0 {
		if _, err = buf.Discard(n); err != nil {
			return buf, n, err
		}
	}

	return buf, n, nil
}

// Read is a short-cut for BOM.Read, where this call returns the initialized BOM, and the result of its BOM.Read call.
func Read(p []byte) (*BOM, int, error) {
	bom := &BOM{}

	n, err := bom.Read(p)

	return bom, n, err
}

// ReadFrom is a short-cut for BOM.ReadFrom, where this call returns the initialized BOM, and the result of its
// BOM.ReadFrom call.
func ReadFrom(r io.Reader) (*BOM, int64, error) {
	bom := &BOM{}

	n, err := bom.ReadFrom(r)

	return bom, n, err
}

func equal[T comparable, S ~[]T](a, b S) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
