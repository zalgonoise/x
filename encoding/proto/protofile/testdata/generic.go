package generic

import (
	"bytes"
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
			byteLen(len(x.Varchar)) +
			len(x.Varchar) +
			byteLen(len(x.ByteSlice)) +
			len(x.ByteSlice) +
			byteLen(x.IntSlice) +
			byteLen(x.EnumField) +
			byteLen(len(x.InnerStruct.Bytes())) +
			len(x.InnerStruct.Bytes()))
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
		e.EncodeInt64(*x.EnumField)
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
