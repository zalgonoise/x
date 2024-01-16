package bom

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var ErrUnsupportedEncoding = errors.New("unsupported encoding")

type BOM struct {
	Unicode   bool
	Size      uint8
	Header    []byte
	ByteOrder binary.ByteOrder
}

func (b *BOM) Read(p []byte) (n int, err error) {
	if len(p) <= 1 {
		return 0, nil
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

	if len(p) <= 2 {
		return 0, nil
	}

	switch {
	case equal(bomUTF8, p[0:3]):
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

	if len(p) <= 3 {
		return 0, nil
	}

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
