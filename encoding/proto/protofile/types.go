package protofile

var (
	Bool    ConcreteType = boolType{}
	Uint32  ConcreteType = uint32Type{}
	Uint64  ConcreteType = uint64Type{}
	Int32   ConcreteType = int32Type{}
	Int64   ConcreteType = int64Type{}
	Float32 ConcreteType = float32Type{}
	Float64 ConcreteType = float64Type{}
	Bytes   ConcreteType = bytesType{}
	String  ConcreteType = stringType{}
)

var protoTypes = map[string]ConcreteType{
	"bool":     Bool,
	"uint32":   Uint32,
	"uint64":   Uint64,
	"sint32":   Int32,
	"sint64":   Int64,
	"int32":    Int32,
	"int64":    Int64,
	"fixed32":  Uint32,
	"fixed64":  Uint64,
	"sfixed32": Int32,
	"sfixed64": Int64,
	"double":   Float32,
	"float":    Float64,
	"string":   String,
	"bytes":    Bytes,
}

var goTypes = map[string]ConcreteType{
	"bool":    Bool,
	"uint32":  Uint32,
	"uint64":  Uint64,
	"int32":   Int32,
	"int64":   Int64,
	"float32": Float32,
	"float64": Float64,
	"string":  String,
	"[]byte":  Bytes,
}

var allocMap = map[ConcreteType]int{
	Bool:    1,
	Int32:   -1,
	Int64:   -1,
	Uint32:  -1,
	Uint64:  -1,
	Float32: 4,
	Float64: 8,
	Bytes:   -1,
	String:  -1,
}

type ConcreteType interface {
	FuncGoString() string
	GoType() string
	WireType() int
	EncoderFunc() string
}

type boolType struct{}

func (boolType) GoType() string      { return "bool" }
func (boolType) WireType() int       { return 0 }
func (boolType) EncoderFunc() string { return "EncodeBool" }

func (boolType) FuncGoString() string {
	return `
// EncodeBool writes the boolean value to the Encoder as a single byte
func (w encoder) EncodeBool(value bool) {
	if value {
		_ = w.b.WriteByte(1)
		return
	}
	_ = w.b.WriteByte(0)
}

`
}

type uint32Type struct{}

func (uint32Type) GoType() string      { return "uint32" }
func (uint32Type) WireType() int       { return 0 }
func (uint32Type) EncoderFunc() string { return "EncodeUint32" }

func (uint32Type) FuncGoString() string {
	return `
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

`
}

type uint64Type struct{}

func (uint64Type) GoType() string      { return "uint64" }
func (uint64Type) WireType() int       { return 0 }
func (uint64Type) EncoderFunc() string { return "EncodeUint64" }

func (uint64Type) FuncGoString() string {
	return `
// EncodeUint64 writes the uint64 value to the Encoder, as a varint
func (w encoder) EncodeUint64(value uint64) int {
	return w.encodeVarint(value)
}

`
}

type int64Type struct{}

func (int64Type) GoType() string      { return "int64" }
func (int64Type) WireType() int       { return 0 }
func (int64Type) EncoderFunc() string { return "EncodeInt64" }

func (int64Type) FuncGoString() string {
	return `
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
`
}

type int32Type struct{}

func (int32Type) GoType() string      { return "int32" }
func (int32Type) WireType() int       { return 0 }
func (int32Type) EncoderFunc() string { return "EncodeInt32" }

func (int32Type) FuncGoString() string {
	return `
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
`
}

type float32Type struct{}

func (float32Type) GoType() string      { return "float32" }
func (float32Type) WireType() int       { return 5 }
func (float32Type) EncoderFunc() string { return "EncodeFloat32" }

func (float32Type) FuncGoString() string {
	return `
// EncodeFloat32 writes the float32 value to the Encoder, as a 4-byte
// buffer
func (w encoder) EncodeFloat32(value float32) int {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, math.Float32bits(value))
	_, _ = w.b.Write(buf)
	return 4
}

`
}

type float64Type struct{}

func (float64Type) GoType() string      { return "float64" }
func (float64Type) WireType() int       { return 1 }
func (float64Type) EncoderFunc() string { return "EncodeFloat64" }

func (float64Type) FuncGoString() string {
	return `
// EncodeFloat64 writes the float64 value to the Encoder, as a 4-byte
// buffer
func (w encoder) EncodeFloat64(value float64) int {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, math.Float64bits(value))
	_, _ = w.b.Write(buf)
	return 8
}

`
}

type bytesType struct{}

func (bytesType) GoType() string      { return "[]byte" }
func (bytesType) WireType() int       { return 2 }
func (bytesType) EncoderFunc() string { return "EncodeBytes" }

func (bytesType) FuncGoString() string {
	return `
// EncodeBytes writes the byte slice to the Encoder, as a length-delimited
// record
func (w encoder) EncodeBytes(value []byte) int {
	n := w.encodeVarint(uint64(len(value)))
	_, _ = w.b.Write(value)
	return n + len(value)
}

`
}

type stringType struct{}

func (stringType) GoType() string      { return "string" }
func (stringType) WireType() int       { return 2 }
func (stringType) EncoderFunc() string { return "EncodeString" }

func (stringType) FuncGoString() string {
	return `
// EncodeString writes the string as a byte slice to the Encoder, 
// as a length-delimited record
func (w encoder) EncodeString(value string) int {
	buf := []byte(value)
	n := w.encodeVarint(uint64(len(buf)))
	_, _ = w.b.Write(buf)
	return n + len(buf)
}

`
}
