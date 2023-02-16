package protofile

import (
	"bytes"
	"fmt"
	"strings"
)

func NewGoFile() *GoFile {
	return &GoFile{
		UniqueTypes:   make(map[string]bool),
		concreteTypes: make(map[string]ConcreteType),
		importsList: map[string]struct{}{
			"bytes":  {},
			"errors": {},
			"io":     {},
		},
	}
}

type GoFile struct {
	Path          string          `json:"path,omitempty"`
	Package       string          `json:"package,omitempty"`
	Types         []*GoType       `json:"types,omitempty"`
	Enums         []*GoEnum       `json:"enums,omitempty"`
	UniqueTypes   map[string]bool `json:"unique_types"`
	concreteTypes map[string]ConcreteType
	importsList   map[string]struct{}
	minAlloc      int
}

func (t GoFile) goStringBuilder() interface {
	String() string
	Bytes() []byte
} {
	buf := new(bytes.Buffer)
	buf.WriteString(t.HeaderGoString())
	for _, typ := range t.Types {
		buf.WriteString(typ.TypeGoString())
		buf.WriteString(typ.EncoderGoString(t))
	}
	for _, enum := range t.Enums {
		buf.WriteString(enum.TypeGoString())
	}

	buf.WriteString(t.EncoderGoString())

	for _, typ := range t.Types {
		buf.WriteString(typ.DecoderGoString(t))
	}
	for _, conc := range t.concreteTypes {
		buf.WriteString(conc.EncoderGoString())
	}
	for _, conc := range t.concreteTypes {
		buf.WriteString(conc.DecoderGoString())
	}

	return buf
}

func (t GoFile) GoBytes() []byte {
	return t.goStringBuilder().Bytes()
}

func (t GoFile) GoString() string {
	return t.goStringBuilder().String()
}

func (g GoFile) EncoderGoString() string {
	return `

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
	for i := 0 ; i < 8 ; i++ {
		v = v >> 8
		if v == 0 {
			return i + i
		}
	}
	return 0
}

`

}

func (g GoFile) HeaderGoString() string {
	sb := new(strings.Builder)

	sb.WriteString(fmt.Sprintf(
		`package %s

import (
`, g.Package))

	for imp := range g.importsList {
		sb.WriteString(fmt.Sprintf(`	"%s"
`, imp))
	}
	sb.WriteString(fmt.Sprintf(
		`)

const (
	minSize = %d
	MaxVarintLen64 = 10
)

`, g.minAlloc))

	for _, typ := range g.Types {
		sb.WriteString("const (\n")
		for _, field := range typ.Fields {
			sb.WriteString(fmt.Sprintf(
				`	header%s%s byte = %d // {%d, %d}
`, typ.Name, field.GoName, field.idAndWire.Header(), field.idAndWire.ID, field.idAndWire.Wire))
		}
		sb.WriteString(")\n")
	}

	return sb.String()
}
