package generic

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

const (
	minSize        = 14
	MaxVarintLen64 = 10
)

const (
	headerShortOk    byte = 8  // {1, 0}
	headerShortValue byte = 18 // {2, 2}
)
const (
	headerGenericBoolField   byte = 8   // {1, 0}
	headerGenericUnsigned32  byte = 16  // {2, 0}
	headerGenericUnsigned64  byte = 24  // {3, 0}
	headerGenericSigned32    byte = 32  // {4, 0}
	headerGenericSigned64    byte = 40  // {5, 0}
	headerGenericInt32       byte = 48  // {6, 0}
	headerGenericInt64       byte = 56  // {7, 0}
	headerGenericFixed32     byte = 64  // {8, 0}
	headerGenericFixed64     byte = 72  // {9, 0}
	headerGenericSfixed32    byte = 80  // {10, 0}
	headerGenericSfixed64    byte = 88  // {11, 0}
	headerGenericFloat32     byte = 101 // {12, 5}
	headerGenericFloat64     byte = 105 // {13, 1}
	headerGenericVarchar     byte = 114 // {14, 2}
	headerGenericByteSlice   byte = 122 // {15, 2}
	headerGenericIntSlice    byte = 128 // {16, 0}
	headerGenericEnumField   byte = 136 // {17, 0}
	headerGenericInnerStruct byte = 146 // {18, 2}
)

// Short describes the message
type Short struct {
	Ok    bool   // id: 1; wire_type: 0
	Value []byte // id: 2; wire_type: 2
}

func (x Short) Bytes() []byte {
	e := newEncoder(
		minSize +
			1 +
			byteLen(len(x.Value)) +
			len(x.Value))
	if x.Ok {
		e.b.WriteByte(8)
		e.EncodeBool(x.Ok)
	}
	if len(x.Value) > 0 {
		e.b.WriteByte(18)
		e.EncodeBytes(x.Value)
	}

	return e.b.Bytes()
}

// Generic describes the message
type Generic struct {
	BoolField   bool     // id: 1; wire_type: 0
	Unsigned32  uint32   // id: 2; wire_type: 0
	Unsigned64  uint64   // id: 3; wire_type: 0
	Signed32    int32    // id: 4; wire_type: 0
	Signed64    int64    // id: 5; wire_type: 0
	Int32       int32    // id: 6; wire_type: 0
	Int64       int64    // id: 7; wire_type: 0
	Fixed32     uint32   // id: 8; wire_type: 0
	Fixed64     uint64   // id: 9; wire_type: 0
	Sfixed32    int32    // id: 10; wire_type: 0
	Sfixed64    int64    // id: 11; wire_type: 0
	Float32     float32  // id: 12; wire_type: 5
	Float64     float64  // id: 13; wire_type: 1
	Varchar     string   // id: 14; wire_type: 2
	ByteSlice   []byte   // id: 15; wire_type: 2
	IntSlice    []uint64 // id: 16; wire_type: 0
	EnumField   *Status  // id: 17; wire_type: 0
	InnerStruct []Short  // id: 18; wire_type: 2
}

func (x Generic) Bytes() []byte {
	e := newEncoder(
		minSize +
			1 +
			byteLen(x.Unsigned32) +
			byteLen(x.Unsigned64) +
			byteLen(x.Signed32) +
			byteLen(x.Signed64) +
			byteLen(x.Int32) +
			byteLen(x.Int64) +
			byteLen(x.Fixed32) +
			byteLen(x.Fixed64) +
			byteLen(x.Sfixed32) +
			byteLen(x.Sfixed64) +
			4 +
			8 +
			byteLen(len(x.Varchar)) +
			len(x.Varchar) +
			byteLen(len(x.ByteSlice)) +
			len(x.ByteSlice) +
			len(x.IntSlice)*8 +
			8 +
			len(x.InnerStruct)*8)
	if x.BoolField {
		e.b.WriteByte(8)
		e.EncodeBool(x.BoolField)
	}
	if x.Unsigned32 != 0 {
		e.b.WriteByte(16)
		e.EncodeUint32(x.Unsigned32)
	}
	if x.Unsigned64 != 0 {
		e.b.WriteByte(24)
		e.EncodeUint64(x.Unsigned64)
	}
	if x.Signed32 != 0 {
		e.b.WriteByte(32)
		e.EncodeInt32(x.Signed32)
	}
	if x.Signed64 != 0 {
		e.b.WriteByte(40)
		e.EncodeInt64(x.Signed64)
	}
	if x.Int32 != 0 {
		e.b.WriteByte(48)
		e.EncodeInt32(x.Int32)
	}
	if x.Int64 != 0 {
		e.b.WriteByte(56)
		e.EncodeInt64(x.Int64)
	}
	if x.Fixed32 != 0 {
		e.b.WriteByte(64)
		e.EncodeUint32(x.Fixed32)
	}
	if x.Fixed64 != 0 {
		e.b.WriteByte(72)
		e.EncodeUint64(x.Fixed64)
	}
	if x.Sfixed32 != 0 {
		e.b.WriteByte(80)
		e.EncodeInt32(x.Sfixed32)
	}
	if x.Sfixed64 != 0 {
		e.b.WriteByte(88)
		e.EncodeInt64(x.Sfixed64)
	}
	if x.Float32 != 0 {
		e.b.WriteByte(101)
		e.EncodeFloat32(x.Float32)
	}
	if x.Float64 != 0 {
		e.b.WriteByte(105)
		e.EncodeFloat64(x.Float64)
	}
	if x.Varchar != "" {
		e.b.WriteByte(114)
		e.EncodeString(x.Varchar)
	}
	if len(x.ByteSlice) > 0 {
		e.b.WriteByte(122)
		e.EncodeBytes(x.ByteSlice)
	}
	for idx := range x.IntSlice {
		if x.IntSlice[idx] != 0 {
			e.b.WriteByte(128)
			e.EncodeUint64(x.IntSlice[idx])
		}
	}
	if x.EnumField != nil {
		e.b.WriteByte(136)
		e.EncodeUint64(uint64(*x.EnumField))
	}
	for idx := range x.InnerStruct {
		if structBytes := x.InnerStruct[idx].Bytes(); len(structBytes) > 0 {
			e.b.WriteByte(146)
			e.EncodeBytes(structBytes)
		}
	}

	return e.b.Bytes()
}

// Status outlines the enumeration
type Status uint64

const (
	NotOk Status = 0
	Ok    Status = 1
)

var (
	convStatusToString = map[Status]string{
		NotOk: "NotOk",
		Ok:    "Ok",
	}

	convStringToStatus = map[string]Status{
		"NotOk": NotOk,
		"Ok":    Ok,
	}
)

func (e Status) String() string {
	return convStatusToString[e]
}

func AsStatus(s string) *Status {
	if v, ok := convStringToStatus[s]; ok {
		return &v
	}
	return nil
}

type encoder struct {
	b *bytes.Buffer
}

func newEncoder(size int) encoder {
	if size == 0 {
		size = minSize
	}
	return encoder{
		b: bytes.NewBuffer(make([]byte, 0, size)),
	}
}

func (w encoder) encodeVarint(value uint64) int {
	i := 0
	for value >= 0x80 {
		_ = w.b.WriteByte(byte(value) | 0x80)
		value >>= 7
		i++
	}
	_ = w.b.WriteByte(byte(value))
	return i + 1
}

type signed interface {
	~int | ~int16 | ~int32 | ~int64
}

type unsigned interface {
	~uint | ~uint16 | ~uint32 | ~uint64
}

type number interface {
	signed | unsigned
}

func byteLen[T number](v T) int {
	for i := 0; i < 8; i++ {
		v = v >> 8
		if v == 0 {
			return i + i
		}
	}
	return 0
}

type decShort struct {
	*bytes.Reader
}

func ToShort(buf []byte) (Short, error) {
	return (&decShort{Reader: bytes.NewReader(buf)}).decode()
}

func (d *decShort) decode() (Short, error) {
	x, err := decodeShort(d.Reader)
	if err != nil && !errors.Is(err, io.EOF) {
		return x, err
	}
	return x, nil
}

func decodeShort(r io.ByteReader) (Short, error) {
	var x = Short{}
	for {
		v, err := r.ReadByte()
		if err != nil {
			return x, err
		}
		switch v {
		case headerShortOk:
			ok, err := decodeBool(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Ok = ok
		case headerShortValue:
			value, err := decodeBytes(r.(io.Reader))
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Value = value
		case 0:
			return x, nil
		default:
			return x, errors.New("invalid header")
		}
	}
}

type decGeneric struct {
	*bytes.Reader
}

func ToGeneric(buf []byte) (Generic, error) {
	return (&decGeneric{Reader: bytes.NewReader(buf)}).decode()
}

func (d *decGeneric) decode() (Generic, error) {
	x, err := decodeGeneric(d.Reader)
	if err != nil && !errors.Is(err, io.EOF) {
		return x, err
	}
	return x, nil
}

func decodeGeneric(r io.ByteReader) (Generic, error) {
	var x = Generic{}
	for {
		v, err := r.ReadByte()
		if err != nil {
			return x, err
		}
		switch v {
		case headerGenericBoolField:
			bool_field, err := decodeBool(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.BoolField = bool_field
		case headerGenericUnsigned32:
			unsigned_32, err := decodeUint32(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Unsigned32 = unsigned_32
		case headerGenericUnsigned64:
			unsigned_64, err := decodeUint64(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Unsigned64 = unsigned_64
		case headerGenericSigned32:
			signed_32, err := decodeInt32(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Signed32 = signed_32
		case headerGenericSigned64:
			signed_64, err := decodeInt64(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Signed64 = signed_64
		case headerGenericInt32:
			int_32, err := decodeInt32(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Int32 = int_32
		case headerGenericInt64:
			int_64, err := decodeInt64(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Int64 = int_64
		case headerGenericFixed32:
			fixed_32, err := decodeUint32(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Fixed32 = fixed_32
		case headerGenericFixed64:
			fixed_64, err := decodeUint64(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Fixed64 = fixed_64
		case headerGenericSfixed32:
			sfixed_32, err := decodeInt32(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Sfixed32 = sfixed_32
		case headerGenericSfixed64:
			sfixed_64, err := decodeInt64(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Sfixed64 = sfixed_64
		case headerGenericFloat32:
			float_32, err := decodeFloat32(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Float32 = float_32
		case headerGenericFloat64:
			float_64, err := decodeFloat64(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Float64 = float_64
		case headerGenericVarchar:
			varchar, err := decodeString(r.(io.Reader))
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.Varchar = varchar
		case headerGenericByteSlice:
			byte_slice, err := decodeBytes(r.(io.Reader))
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.ByteSlice = byte_slice
		case headerGenericIntSlice:
			int_slice, err := decodeUint64(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.IntSlice = append(x.IntSlice, int_slice)
		case headerGenericEnumField:
			enum_field, err := decodeUint64(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.EnumField = (*Status)(&enum_field)
		case headerGenericInnerStruct:
			_, err := r.ReadByte() // length byte
			if err != nil {
				return x, err
			}
			inner_struct, err := decodeShort(r)
			if err != nil && !errors.Is(err, io.EOF) {
				return x, err
			}
			x.InnerStruct = append(x.InnerStruct, inner_struct)
		case 0:
			return x, nil
		default:
			return x, errors.New("invalid header")
		}
	}
}

// EncodeBool writes the boolean value to the Encoder as a single byte
func (w encoder) EncodeBool(value bool) {
	if value {
		_ = w.b.WriteByte(1)
		return
	}
	_ = w.b.WriteByte(0)
}

// EncodeUint64 writes the uint64 value to the Encoder, as a varint
func (w encoder) EncodeUint64(value uint64) int {
	return w.encodeVarint(value)
}

// EncodeFloat32 writes the float32 value to the Encoder, as a 4-byte
// buffer
func (w encoder) EncodeFloat32(value float32) int {
	var buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(value))
	_, _ = w.b.Write(buf)
	return 4
}

// EncodeString writes the string as a byte slice to the Encoder,
// as a length-delimited record
func (w encoder) EncodeString(value string) int {
	buf := []byte(value)
	n := w.encodeVarint(uint64(len(buf)))
	_, _ = w.b.Write(buf)
	return n + len(buf)
}

// EncodeBytes writes the byte slice to the Encoder, as a length-delimited
// record
func (w encoder) EncodeBytes(value []byte) int {
	n := w.encodeVarint(uint64(len(value)))
	_, _ = w.b.Write(value)
	return n + len(value)
}

// EncodeUint32 writes the uint32 value to the Encoder, as a varint
func (w encoder) EncodeUint32(value uint32) int {
	i := 0
	for value >= 0x80 {
		_ = w.b.WriteByte(byte(value) | 0x80)
		value >>= 7
		i++
	}
	_ = w.b.WriteByte(byte(value))
	return i + 1
}

// EncodeInt32 writes the int32 value to the Encoder, as a zig-zag
// encoded varint
func (w encoder) EncodeInt32(value int32) int {
	v := uint32((value << 1) ^ (value >> 31))
	i := 0
	for v >= 0x80 {
		_ = w.b.WriteByte(byte(v) | 0x80)
		v >>= 7
		i++
	}
	_ = w.b.WriteByte(byte(v))
	return i + 1
}

// EncodeInt64 writes the int64 value to the Encoder, as a zig-zag
// encoded varint
func (w encoder) EncodeInt64(value int64) int {
	v := uint64((value << 1) ^ (value >> 63))
	i := 0
	for v >= 0x80 {
		_ = w.b.WriteByte(byte(v) | 0x80)
		v >>= 7
		i++
	}
	_ = w.b.WriteByte(byte(v))
	return i + 1
}

// EncodeFloat64 writes the float64 value to the Encoder, as a 4-byte
// buffer
func (w encoder) EncodeFloat64(value float64) int {
	var buf = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, math.Float64bits(value))
	_, _ = w.b.Write(buf)
	return 8
}

func decodeFloat32(r io.ByteReader) (v float32, err error) {
	var arr [4]byte
	for idx := range arr {
		arr[idx], err = r.ReadByte()
		if err != nil {
			return v, err
		}
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(arr[:])), nil
}

func decodeString(r io.Reader) (string, error) {
	length, err := decodeUint64(r.(io.ByteReader))
	if err != nil {
		return "", err
	}
	data := make([]byte, length)
	_, err = r.Read(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodeBool(r io.ByteReader) (bool, error) {
	var x bool
	byt, err := r.ReadByte()
	if err != nil {
		return x, err
	}
	if byt == 1 {
		return true, nil
	}
	return false, nil
}

func decodeUint64(r io.ByteReader) (uint64, error) {
	var x uint64
	var s uint
	var i int
	for {
		byt, err := r.ReadByte()
		if err != nil {
			return x, err
		}
		i++
		if i == MaxVarintLen64 {
			return 0, errors.New("varint overflow") // overflow
		}
		if byt < 0x80 {
			if i == MaxVarintLen64-1 && byt > 1 {
				return 0, errors.New("varint overflow") // overflow
			}
			return x | uint64(byt)<<s, nil
		}
		x |= uint64(byt&0x7f) << s
		s += 7
	}
}

func decodeInt32(r io.ByteReader) (int32, error) {
	var x uint32
	var s uint
	var i int
	for {
		byt, err := r.ReadByte()
		if err != nil {
			return int32(x), err
		}
		i++
		if i == MaxVarintLen64 {
			return 0, errors.New("varint overflow") // overflow
		}
		if byt < 0x80 {
			if i == MaxVarintLen64-1 && byt > 1 {
				return 0, errors.New("varint overflow") // overflow
			}
			n := x | uint32(byt)<<s
			return int32((n >> 1) ^ -(n & 1)), nil
		}
		x |= uint32(byt&0x7f) << s
		s += 7
	}
}

func decodeInt64(r io.ByteReader) (int64, error) {
	var x uint64
	var s uint
	var i int
	for {
		byt, err := r.ReadByte()
		if err != nil {
			return int64(x), err
		}
		i++
		if i == MaxVarintLen64 {
			return 0, errors.New("varint overflow") // overflow
		}
		if byt < 0x80 {
			if i == MaxVarintLen64-1 && byt > 1 {
				return 0, errors.New("varint overflow") // overflow
			}
			n := x | uint64(byt)<<s
			return int64((n >> 1) ^ -(n & 1)), nil
		}
		x |= uint64(byt&0x7f) << s
		s += 7
	}
}

func decodeFloat64(r io.ByteReader) (v float64, err error) {
	var arr [8]byte
	for idx := range arr {
		arr[idx], err = r.ReadByte()
		if err != nil {
			return v, err
		}
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(arr[:])), nil
}

func decodeBytes(r io.Reader) ([]byte, error) {
	length, err := decodeUint64(r.(io.ByteReader))
	if err != nil {
		return nil, err
	}
	data := make([]byte, length)
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func decodeUint32(r io.ByteReader) (uint32, error) {
	var x uint32
	var s uint
	var i int
	for {
		byt, err := r.ReadByte()
		if err != nil {
			return x, err
		}
		i++
		if i == MaxVarintLen64 {
			return 0, errors.New("varint overflow") // overflow
		}
		if byt < 0x80 {
			if i == MaxVarintLen64-1 && byt > 1 {
				return 0, errors.New("varint overflow") // overflow
			}
			return x | uint32(byt)<<s, nil
		}
		x |= uint32(byt&0x7f) << s
		s += 7
	}
}
