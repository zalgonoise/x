package generic

import (
	"bytes"
	"encoding/binary"
	"math"
)

const minSize = 14

// Short describes the message
type Short struct {
	Ok    bool   // id:1; wire type:0
	Value []byte // id:2; wire type:2
}

func (x Short) Bytes() []byte {
	e := newEncoder(
		minSize +
			1 +
			byteLen(len(x.Value)) +
			len(x.Value))
	e.b.WriteByte(8)
	e.EncodeBool(x.Ok)
	e.b.WriteByte(18)
	e.EncodeBytes(x.Value)

	return e.b.Bytes()
}

// Generic describes the message
type Generic struct {
	BoolField   bool     // id:1; wire type:0
	Unsigned32  uint32   // id:2; wire type:0
	Unsigned64  uint64   // id:3; wire type:0
	Signed32    int32    // id:4; wire type:0
	Signed64    int64    // id:5; wire type:0
	Int32       int32    // id:6; wire type:0
	Int64       int64    // id:7; wire type:0
	Fixed32     uint32   // id:8; wire type:0
	Fixed64     uint64   // id:9; wire type:0
	Sfixed32    int32    // id:10; wire type:0
	Sfixed64    int64    // id:11; wire type:0
	Float32     float32  // id:12; wire type:5
	Float64     float64  // id:13; wire type:1
	Varchar     string   // id:14; wire type:2
	ByteSlice   []byte   // id:15; wire type:2
	IntSlice    []uint64 // id:16; wire type:0
	EnumField   *Status  // id:17; wire type:2
	InnerStruct []Short  // id:18; wire type:2
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
			8 +
			8 +
			byteLen(len(x.Varchar)) +
			len(x.Varchar) +
			byteLen(len(x.ByteSlice)) +
			len(x.ByteSlice) +
			len(x.IntSlice)*8 +
			8 +
			len(x.InnerStruct)*8)
	e.b.WriteByte(8)
	e.EncodeBool(x.BoolField)
	e.b.WriteByte(16)
	e.EncodeUint32(x.Unsigned32)
	e.b.WriteByte(24)
	e.EncodeUint64(x.Unsigned64)
	e.b.WriteByte(32)
	e.EncodeInt32(x.Signed32)
	e.b.WriteByte(40)
	e.EncodeInt64(x.Signed64)
	e.b.WriteByte(48)
	e.EncodeInt32(x.Int32)
	e.b.WriteByte(56)
	e.EncodeInt64(x.Int64)
	e.b.WriteByte(64)
	e.EncodeUint32(x.Fixed32)
	e.b.WriteByte(72)
	e.EncodeUint64(x.Fixed64)
	e.b.WriteByte(80)
	e.EncodeInt32(x.Sfixed32)
	e.b.WriteByte(88)
	e.EncodeInt64(x.Sfixed64)
	e.b.WriteByte(101)
	e.EncodeFloat32(x.Float32)
	e.b.WriteByte(105)
	e.EncodeFloat64(x.Float64)
	e.b.WriteByte(114)
	e.EncodeString(x.Varchar)
	e.b.WriteByte(122)
	e.EncodeBytes(x.ByteSlice)
	for idx := range x.IntSlice {
		e.b.WriteByte(128)
		e.EncodeUint64(x.IntSlice[idx])
	}
	if x.EnumField != nil {
		e.b.WriteByte(138)
		e.EncodeInt64(int64(*x.EnumField))
	}
	for idx := range x.InnerStruct {
		e.b.WriteByte(146)
		e.EncodeBytes(x.InnerStruct[idx].Bytes())
	}

	return e.b.Bytes()
}

// Status outlines the enumeration
type Status int64

const (
	NotOk Status = 0
	Ok    Status = 1
)

var convStatusToString = map[Status]string{
	NotOk: "NotOk",
	Ok:    "Ok",
}

var convStringToStatus = map[string]Status{
	"NotOk": NotOk,
	"Ok":    Ok,
}

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

func byteLen[T number](v T) (size int) {
	for i := 0; i < 8; i++ {
		v = v >> 8
		if v == 0 {
			return i + i
		}
	}
	return 0
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

// EncodeFloat32 writes the float32 value to the Encoder, as a 4-byte
// buffer
func (w encoder) EncodeFloat32(value float32) int {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, math.Float32bits(value))
	_, _ = w.b.Write(buf)
	return 4
}

// EncodeBool writes the boolean value to the Encoder as a single byte
func (w encoder) EncodeBool(value bool) {
	if value {
		_ = w.b.WriteByte(1)
		return
	}
	_ = w.b.WriteByte(0)
}

// EncodeBytes writes the byte slice to the Encoder, as a length-delimited
// record
func (w encoder) EncodeBytes(value []byte) int {
	n := w.encodeVarint(uint64(len(value)))
	_, _ = w.b.Write(value)
	return n + len(value)
}

// EncodeFloat64 writes the float64 value to the Encoder, as a 4-byte
// buffer
func (w encoder) EncodeFloat64(value float64) int {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, math.Float64bits(value))
	_, _ = w.b.Write(buf)
	return 8
}

// EncodeString writes the string as a byte slice to the Encoder,
// as a length-delimited record
func (w encoder) EncodeString(value string) int {
	buf := []byte(value)
	n := w.encodeVarint(uint64(len(buf)))
	_, _ = w.b.Write(buf)
	return n + len(buf)
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

// EncodeUint64 writes the uint64 value to the Encoder, as a varint
func (w encoder) EncodeUint64(value uint64) int {
	return w.encodeVarint(value)
}
