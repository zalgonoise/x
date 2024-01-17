package bom

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

func TestBOM_Read(t *testing.T) {
	data := []byte("data")
	shortData := []byte("dat")
	shorterData := []byte("da")

	for _, testcase := range []struct {
		name     string
		input    []byte
		numBytes int
		err      error

		isUnicode bool
		size      uint8
		header    []byte
		byteOrder binary.ByteOrder
	}{
		{
			name:  "NoBOM",
			input: data,
		},
		{
			name:  "NoBOM/ShortData",
			input: shortData,
		},
		{
			name:  "NoBOM/ShorterData",
			input: shorterData,
		},
		{
			name:      "BOM/UTF8",
			input:     append(bomUTF8, data...),
			numBytes:  3,
			isUnicode: true,
			size:      8,
			header:    bomUTF8,
		},
		{
			name:      "BOM/UTF16BE",
			input:     append(bomUTF16BE, data...),
			numBytes:  2,
			isUnicode: true,
			size:      16,
			header:    bomUTF16BE,
			byteOrder: binary.BigEndian,
		},
		{
			name:      "BOM/UTF16LE",
			input:     append(bomUTF16LE, data...),
			numBytes:  2,
			isUnicode: true,
			size:      16,
			header:    bomUTF16LE,
			byteOrder: binary.LittleEndian,
		},
		{
			name:      "BOM/UTF32BE",
			input:     append(bomUTF32BE, data...),
			numBytes:  4,
			isUnicode: true,
			size:      32,
			header:    bomUTF32BE,
			byteOrder: binary.BigEndian,
		},
		{
			name:      "BOM/UTF32LE",
			input:     append(bomUTF32LE, data...),
			numBytes:  4,
			isUnicode: true,
			size:      32,
			header:    bomUTF32LE,
			byteOrder: binary.LittleEndian,
		},
		{
			name:      "BOM/UTF7",
			input:     append(bomUTF7, data...),
			numBytes:  3,
			isUnicode: true,
			header:    bomUTF7,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/UTF1",
			input:     append(bomUTF1, data...),
			numBytes:  3,
			isUnicode: true,
			header:    bomUTF1,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/UTFEBCDIC",
			input:     append(bomUTFEBCDIC, data...),
			numBytes:  4,
			isUnicode: true,
			header:    bomUTFEBCDIC,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/SCSU",
			input:     append(bomSCSU, data...),
			numBytes:  3,
			isUnicode: true,
			header:    bomSCSU,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/BOCU1",
			input:     append(bomBOCU1, data...),
			numBytes:  3,
			isUnicode: true,
			header:    bomBOCU1,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/GB18030",
			input:     append(bomGB18030, data...),
			numBytes:  4,
			isUnicode: true,
			header:    bomGB18030,
			err:       ErrUnsupportedEncoding,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			bom := &BOM{}

			n, err := bom.Read(testcase.input)
			isEqual(t, testcase.numBytes, n)
			isEqual(t, true, errors.Is(err, testcase.err))

			isEqual(t, testcase.isUnicode, bom.Unicode)
			isEqual(t, testcase.size, bom.Size)
			isEqual(t, len(testcase.header), len(bom.Header))
			for i := range testcase.header {
				isEqual(t, testcase.header[i], bom.Header[i])
			}
			isEqual(t, testcase.byteOrder, bom.ByteOrder)
		})
	}
}

func TestBOM_ReadFrom(t *testing.T) {
	data := []byte("data")
	shortData := []byte("dat")
	shorterData := []byte("da")

	for _, testcase := range []struct {
		name     string
		input    io.Reader
		numBytes int64
		err      error

		isUnicode bool
		size      uint8
		header    []byte
		byteOrder binary.ByteOrder
	}{
		{
			name:  "NoBOM",
			input: bytes.NewReader(data),
		},
		{
			name:  "NoBOM/ShortData",
			input: bytes.NewReader(shortData),
			err:   io.EOF,
		},
		{
			name:  "NoBOM/ShorterData",
			input: bytes.NewReader(shorterData),
			err:   io.EOF,
		},
		{
			name:      "BOM/UTF8",
			input:     bytes.NewReader(append(bomUTF8, data...)),
			numBytes:  3,
			isUnicode: true,
			size:      8,
			header:    bomUTF8,
		},
		{
			name:      "BOM/UTF16BE",
			input:     bytes.NewReader(append(bomUTF16BE, data...)),
			numBytes:  2,
			isUnicode: true,
			size:      16,
			header:    bomUTF16BE,
			byteOrder: binary.BigEndian,
		},
		{
			name:      "BOM/UTF16LE",
			input:     bytes.NewReader(append(bomUTF16LE, data...)),
			numBytes:  2,
			isUnicode: true,
			size:      16,
			header:    bomUTF16LE,
			byteOrder: binary.LittleEndian,
		},
		{
			name:      "BOM/UTF32BE",
			input:     bytes.NewReader(append(bomUTF32BE, data...)),
			numBytes:  4,
			isUnicode: true,
			size:      32,
			header:    bomUTF32BE,
			byteOrder: binary.BigEndian,
		},
		{
			name:      "BOM/UTF32LE",
			input:     bytes.NewReader(append(bomUTF32LE, data...)),
			numBytes:  4,
			isUnicode: true,
			size:      32,
			header:    bomUTF32LE,
			byteOrder: binary.LittleEndian,
		},
		{
			name:      "BOM/UTF7",
			input:     bytes.NewReader(append(bomUTF7, data...)),
			numBytes:  3,
			isUnicode: true,
			header:    bomUTF7,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/UTF1",
			input:     bytes.NewReader(append(bomUTF1, data...)),
			numBytes:  3,
			isUnicode: true,
			header:    bomUTF1,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/UTFEBCDIC",
			input:     bytes.NewReader(append(bomUTFEBCDIC, data...)),
			numBytes:  4,
			isUnicode: true,
			header:    bomUTFEBCDIC,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/SCSU",
			input:     bytes.NewReader(append(bomSCSU, data...)),
			numBytes:  3,
			isUnicode: true,
			header:    bomSCSU,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/BOCU1",
			input:     bytes.NewReader(append(bomBOCU1, data...)),
			numBytes:  3,
			isUnicode: true,
			header:    bomBOCU1,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/GB18030",
			input:     bytes.NewReader(append(bomGB18030, data...)),
			numBytes:  4,
			isUnicode: true,
			header:    bomGB18030,
			err:       ErrUnsupportedEncoding,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			bom := &BOM{}

			n, err := bom.ReadFrom(testcase.input)
			t.Log(err)
			isEqual(t, testcase.numBytes, n)
			isEqual(t, true, errors.Is(err, testcase.err))

			isEqual(t, testcase.isUnicode, bom.Unicode)
			isEqual(t, testcase.size, bom.Size)
			isEqual(t, len(testcase.header), len(bom.Header))
			for i := range testcase.header {
				isEqual(t, testcase.header[i], bom.Header[i])
			}
			isEqual(t, testcase.byteOrder, bom.ByteOrder)
		})
	}
}

func TestDrop(t *testing.T) {
	data := []byte("data")
	shortData := []byte("dat")
	shorterData := []byte("da")

	for _, testcase := range []struct {
		name     string
		input    []byte
		wants    []byte
		numBytes int
		err      error
	}{
		{
			name:  "NoBOM",
			input: data,
			wants: data,
		},
		{
			name:  "NoBOM/ShortData",
			input: shortData,
			wants: shortData,
		},
		{
			name:  "NoBOM/ShorterData",
			input: shorterData,
			wants: shorterData,
		},
		{
			name:     "BOM/UTF8",
			input:    append(bomUTF8, data...),
			wants:    data,
			numBytes: 3,
		},
		{
			name:     "BOM/UTF16BE",
			input:    append(bomUTF16BE, data...),
			wants:    data,
			numBytes: 2,
		},
		{
			name:     "BOM/UTF16LE",
			input:    append(bomUTF16LE, data...),
			wants:    data,
			numBytes: 2,
		},
		{
			name:     "BOM/UTF32BE",
			input:    append(bomUTF32BE, data...),
			wants:    data,
			numBytes: 4,
		},
		{
			name:     "BOM/UTF32LE",
			input:    append(bomUTF32LE, data...),
			wants:    data,
			numBytes: 4,
		},
		{
			name:     "BOM/UTF7",
			input:    append(bomUTF7, data...),
			wants:    append(bomUTF7, data...),
			numBytes: 3,
			err:      ErrUnsupportedEncoding,
		},
		{
			name:     "BOM/UTF1",
			input:    append(bomUTF1, data...),
			wants:    append(bomUTF1, data...),
			numBytes: 3,
			err:      ErrUnsupportedEncoding,
		},
		{
			name:     "BOM/UTFEBCDIC",
			input:    append(bomUTFEBCDIC, data...),
			wants:    append(bomUTFEBCDIC, data...),
			numBytes: 4,
			err:      ErrUnsupportedEncoding,
		},
		{
			name:     "BOM/SCSU",
			input:    append(bomSCSU, data...),
			wants:    append(bomSCSU, data...),
			numBytes: 3,
			err:      ErrUnsupportedEncoding,
		},
		{
			name:     "BOM/BOCU1",
			input:    append(bomBOCU1, data...),
			wants:    append(bomBOCU1, data...),
			numBytes: 3,
			err:      ErrUnsupportedEncoding,
		},
		{
			name:     "BOM/GB18030",
			input:    append(bomGB18030, data...),
			wants:    append(bomGB18030, data...),
			numBytes: 4,
			err:      ErrUnsupportedEncoding,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			out, n, err := Drop(testcase.input)
			isEqual(t, testcase.numBytes, n)
			isEqual(t, true, errors.Is(err, testcase.err))

			isEqual(t, len(testcase.wants), len(out))
			for i := range testcase.wants {
				isEqual(t, testcase.wants[i], out[i])
			}
		})
	}
}

func TestDiscard(t *testing.T) {
	data := []byte("data")
	shortData := []byte("dat")
	shorterData := []byte("da")

	for _, testcase := range []struct {
		name     string
		input    io.Reader
		wants    io.Reader
		numBytes int
		err      error
	}{
		{
			name:  "NoBOM",
			input: bytes.NewReader(data),
			wants: bytes.NewReader(data),
		},
		{
			name:  "NoBOM/ShortData",
			input: bytes.NewReader(shortData),
			wants: bytes.NewReader(shortData),
			err:   io.EOF,
		},
		{
			name:  "NoBOM/ShorterData",
			input: bytes.NewReader(shorterData),
			wants: bytes.NewReader(shorterData),
			err:   io.EOF,
		},
		{
			name:     "BOM/UTF8",
			input:    bytes.NewReader(append(bomUTF8, data...)),
			wants:    bytes.NewReader(data),
			numBytes: 3,
		},
		{
			name:     "BOM/UTF16BE",
			input:    bytes.NewReader(append(bomUTF16BE, data...)),
			wants:    bytes.NewReader(data),
			numBytes: 2,
		},
		{
			name:     "BOM/UTF16LE",
			input:    bytes.NewReader(append(bomUTF16LE, data...)),
			wants:    bytes.NewReader(data),
			numBytes: 2,
		},
		{
			name:     "BOM/UTF32BE",
			input:    bytes.NewReader(append(bomUTF32BE, data...)),
			wants:    bytes.NewReader(data),
			numBytes: 4,
		},
		{
			name:     "BOM/UTF32LE",
			input:    bytes.NewReader(append(bomUTF32LE, data...)),
			wants:    bytes.NewReader(data),
			numBytes: 4,
		},
		{
			name:     "BOM/UTF7",
			input:    bytes.NewReader(append(bomUTF7, data...)),
			wants:    bytes.NewReader(append(bomUTF7, data...)),
			numBytes: 3,
			err:      ErrUnsupportedEncoding,
		},
		{
			name:     "BOM/UTF1",
			input:    bytes.NewReader(append(bomUTF1, data...)),
			wants:    bytes.NewReader(append(bomUTF1, data...)),
			numBytes: 3,
			err:      ErrUnsupportedEncoding,
		},
		{
			name:     "BOM/UTFEBCDIC",
			input:    bytes.NewReader(append(bomUTFEBCDIC, data...)),
			wants:    bytes.NewReader(append(bomUTFEBCDIC, data...)),
			numBytes: 4,
			err:      ErrUnsupportedEncoding,
		},
		{
			name:     "BOM/SCSU",
			input:    bytes.NewReader(append(bomSCSU, data...)),
			wants:    bytes.NewReader(append(bomSCSU, data...)),
			numBytes: 3,
			err:      ErrUnsupportedEncoding,
		},
		{
			name:     "BOM/BOCU1",
			input:    bytes.NewReader(append(bomBOCU1, data...)),
			wants:    bytes.NewReader(append(bomBOCU1, data...)),
			numBytes: 3,
			err:      ErrUnsupportedEncoding,
		},
		{
			name:     "BOM/GB18030",
			input:    bytes.NewReader(append(bomGB18030, data...)),
			wants:    bytes.NewReader(append(bomGB18030, data...)),
			numBytes: 4,
			err:      ErrUnsupportedEncoding,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			out, n, err := Discard(testcase.input)
			isEqual(t, testcase.numBytes, n)
			isEqual(t, true, errors.Is(err, testcase.err))

			wants, err := io.ReadAll(testcase.wants)
			isEqual(t, nil, err)

			got, err := io.ReadAll(out)
			isEqual(t, nil, err)

			isEqual(t, len(wants), len(got))
			for i := range wants {
				isEqual(t, wants[i], got[i])
			}
		})
	}
}

func TestRead(t *testing.T) {
	data := []byte("data")
	shortData := []byte("dat")
	shorterData := []byte("da")

	for _, testcase := range []struct {
		name     string
		input    []byte
		numBytes int
		err      error

		isUnicode bool
		size      uint8
		header    []byte
		byteOrder binary.ByteOrder
	}{
		{
			name:  "NoBOM",
			input: data,
		},
		{
			name:  "NoBOM/ShortData",
			input: shortData,
		},
		{
			name:  "NoBOM/ShorterData",
			input: shorterData,
		},
		{
			name:      "BOM/UTF8",
			input:     append(bomUTF8, data...),
			numBytes:  3,
			isUnicode: true,
			size:      8,
			header:    bomUTF8,
		},
		{
			name:      "BOM/UTF16BE",
			input:     append(bomUTF16BE, data...),
			numBytes:  2,
			isUnicode: true,
			size:      16,
			header:    bomUTF16BE,
			byteOrder: binary.BigEndian,
		},
		{
			name:      "BOM/UTF16LE",
			input:     append(bomUTF16LE, data...),
			numBytes:  2,
			isUnicode: true,
			size:      16,
			header:    bomUTF16LE,
			byteOrder: binary.LittleEndian,
		},
		{
			name:      "BOM/UTF32BE",
			input:     append(bomUTF32BE, data...),
			numBytes:  4,
			isUnicode: true,
			size:      32,
			header:    bomUTF32BE,
			byteOrder: binary.BigEndian,
		},
		{
			name:      "BOM/UTF32LE",
			input:     append(bomUTF32LE, data...),
			numBytes:  4,
			isUnicode: true,
			size:      32,
			header:    bomUTF32LE,
			byteOrder: binary.LittleEndian,
		},
		{
			name:      "BOM/UTF7",
			input:     append(bomUTF7, data...),
			numBytes:  3,
			isUnicode: true,
			header:    bomUTF7,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/UTF1",
			input:     append(bomUTF1, data...),
			numBytes:  3,
			isUnicode: true,
			header:    bomUTF1,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/UTFEBCDIC",
			input:     append(bomUTFEBCDIC, data...),
			numBytes:  4,
			isUnicode: true,
			header:    bomUTFEBCDIC,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/SCSU",
			input:     append(bomSCSU, data...),
			numBytes:  3,
			isUnicode: true,
			header:    bomSCSU,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/BOCU1",
			input:     append(bomBOCU1, data...),
			numBytes:  3,
			isUnicode: true,
			header:    bomBOCU1,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/GB18030",
			input:     append(bomGB18030, data...),
			numBytes:  4,
			isUnicode: true,
			header:    bomGB18030,
			err:       ErrUnsupportedEncoding,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			bom, n, err := Read(testcase.input)
			isEqual(t, testcase.numBytes, n)
			isEqual(t, true, errors.Is(err, testcase.err))

			isEqual(t, testcase.isUnicode, bom.Unicode)
			isEqual(t, testcase.size, bom.Size)
			isEqual(t, len(testcase.header), len(bom.Header))
			for i := range testcase.header {
				isEqual(t, testcase.header[i], bom.Header[i])
			}
			isEqual(t, testcase.byteOrder, bom.ByteOrder)
		})
	}
}

func TestReadFrom(t *testing.T) {
	data := []byte("data")
	shortData := []byte("dat")
	shorterData := []byte("da")

	for _, testcase := range []struct {
		name     string
		input    io.Reader
		numBytes int64
		err      error

		isUnicode bool
		size      uint8
		header    []byte
		byteOrder binary.ByteOrder
	}{
		{
			name:  "NoBOM",
			input: bytes.NewReader(data),
		},
		{
			name:  "NoBOM/ShortData",
			input: bytes.NewReader(shortData),
			err:   io.EOF,
		},
		{
			name:  "NoBOM/ShorterData",
			input: bytes.NewReader(shorterData),
			err:   io.EOF,
		},
		{
			name:      "BOM/UTF8",
			input:     bytes.NewReader(append(bomUTF8, data...)),
			numBytes:  3,
			isUnicode: true,
			size:      8,
			header:    bomUTF8,
		},
		{
			name:      "BOM/UTF16BE",
			input:     bytes.NewReader(append(bomUTF16BE, data...)),
			numBytes:  2,
			isUnicode: true,
			size:      16,
			header:    bomUTF16BE,
			byteOrder: binary.BigEndian,
		},
		{
			name:      "BOM/UTF16LE",
			input:     bytes.NewReader(append(bomUTF16LE, data...)),
			numBytes:  2,
			isUnicode: true,
			size:      16,
			header:    bomUTF16LE,
			byteOrder: binary.LittleEndian,
		},
		{
			name:      "BOM/UTF32BE",
			input:     bytes.NewReader(append(bomUTF32BE, data...)),
			numBytes:  4,
			isUnicode: true,
			size:      32,
			header:    bomUTF32BE,
			byteOrder: binary.BigEndian,
		},
		{
			name:      "BOM/UTF32LE",
			input:     bytes.NewReader(append(bomUTF32LE, data...)),
			numBytes:  4,
			isUnicode: true,
			size:      32,
			header:    bomUTF32LE,
			byteOrder: binary.LittleEndian,
		},
		{
			name:      "BOM/UTF7",
			input:     bytes.NewReader(append(bomUTF7, data...)),
			numBytes:  3,
			isUnicode: true,
			header:    bomUTF7,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/UTF1",
			input:     bytes.NewReader(append(bomUTF1, data...)),
			numBytes:  3,
			isUnicode: true,
			header:    bomUTF1,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/UTFEBCDIC",
			input:     bytes.NewReader(append(bomUTFEBCDIC, data...)),
			numBytes:  4,
			isUnicode: true,
			header:    bomUTFEBCDIC,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/SCSU",
			input:     bytes.NewReader(append(bomSCSU, data...)),
			numBytes:  3,
			isUnicode: true,
			header:    bomSCSU,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/BOCU1",
			input:     bytes.NewReader(append(bomBOCU1, data...)),
			numBytes:  3,
			isUnicode: true,
			header:    bomBOCU1,
			err:       ErrUnsupportedEncoding,
		},
		{
			name:      "BOM/GB18030",
			input:     bytes.NewReader(append(bomGB18030, data...)),
			numBytes:  4,
			isUnicode: true,
			header:    bomGB18030,
			err:       ErrUnsupportedEncoding,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			bom, n, err := ReadFrom(testcase.input)
			t.Log(err)
			isEqual(t, testcase.numBytes, n)
			isEqual(t, true, errors.Is(err, testcase.err))

			isEqual(t, testcase.isUnicode, bom.Unicode)
			isEqual(t, testcase.size, bom.Size)
			isEqual(t, len(testcase.header), len(bom.Header))
			for i := range testcase.header {
				isEqual(t, testcase.header[i], bom.Header[i])
			}
			isEqual(t, testcase.byteOrder, bom.ByteOrder)
		})
	}
}

func isEqual[T comparable](t *testing.T, wants, got T) {
	if got != wants {
		t.Errorf("output mismatch error: wanted %v ; got %v", wants, got)
		t.Fail()

		return
	}

	t.Logf("output matched expected value: %v", wants)
}
