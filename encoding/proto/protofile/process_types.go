package protofile

import (
	"strconv"
	"strings"
)

type GoField struct {
	IsRepeated bool   `json:"is_repeated,omitempty"`
	IsOptional bool   `json:"is_optional,omitempty"`
	IsStruct   bool   `json:"is_struct,omitempty"`
	GoType     string `json:"go_type,omitempty"`
	GoName     string `json:"go_name,omitempty"`
	ProtoType  string `json:"proto_type,omitempty"`
	ProtoIndex int    `json:"proto_index"`
	ProtoName  string `json:"proto_name,omitempty"`
	idAndWire  IDAndWire
}

type GoType struct {
	Name        string    `json:"name,omitempty"`
	Fields      []GoField `json:"fields,omitempty"`
	uniqueNames map[string]struct{}
	uniqueIDs   map[int]struct{}
}

func (t GoType) GoString() string {
	sb := new(strings.Builder)
	sb.WriteString("\n\n// ")
	sb.WriteString(t.Name)
	sb.WriteString(" describes the message")
	sb.WriteString("\ntype ")
	sb.WriteString(t.Name)
	sb.WriteString(" struct {\n")
	for _, f := range t.Fields {
		sb.WriteByte('\t')
		sb.WriteString(f.GoName)
		sb.WriteByte('\t')
		if f.IsRepeated {
			sb.WriteString("[]")
		}
		if f.IsOptional {
			sb.WriteString("*")
		}
		sb.WriteString(f.GoType)
		sb.WriteByte('\t')
		sb.WriteString("// id:")
		sb.WriteString(strconv.Itoa(f.ProtoIndex))
		sb.WriteString("; wire type:")
		sb.WriteString(strconv.Itoa(f.idAndWire.Wire))
		sb.WriteByte('\n')
	}
	sb.WriteString("}\n")

	return sb.String()
}

type EnumField struct {
	Index     int    `json:"index"`
	GoName    string `json:"go_name,omitempty"`
	ProtoName string `json:"proto_name,omitempty"`
}

type GoEnum struct {
	ProtoName   string      `json:"proto_name,omitempty"`
	GoName      string      `json:"go_name,omitempty"`
	Fields      []EnumField `json:"fields,omitempty"`
	uniqueNames map[string]struct{}
	uniqueIDs   map[int]struct{}
}

func (t GoEnum) GoString() string {
	sb := new(strings.Builder)
	sb.WriteString("\n\n// ")
	sb.WriteString(t.GoName)
	sb.WriteString(" outlines the enumeration")
	sb.WriteString("\ntype ")
	sb.WriteString(t.GoName)
	sb.WriteString(" int64\n\n")

	sb.WriteString("const (\n")
	for _, f := range t.Fields {
		sb.WriteByte('\t')
		sb.WriteString(f.GoName)
		sb.WriteByte('\t')
		sb.WriteString(t.GoName)
		sb.WriteString(" = ")
		sb.WriteString(strconv.Itoa(f.Index))
		sb.WriteByte('\n')
	}
	sb.WriteString(")\n\n")
	sb.WriteString("var conv")
	sb.WriteString(t.GoName)
	sb.WriteString("ToString = map[")
	sb.WriteString(t.GoName)
	sb.WriteString("]string{\n")

	for _, f := range t.Fields {
		sb.WriteByte('\t')
		sb.WriteString(f.GoName)
		sb.WriteString(`: "`)
		sb.WriteString(f.GoName)
		sb.WriteString(`",`)
		sb.WriteByte('\n')
	}
	sb.WriteString("}\n\n")
	sb.WriteString("var convStringTo")
	sb.WriteString(t.GoName)
	sb.WriteString(" = map[string]")
	sb.WriteString(t.GoName)
	sb.WriteString("{\n")

	for _, f := range t.Fields {
		sb.WriteByte('\t')
		sb.WriteByte('"')
		sb.WriteString(f.GoName)
		sb.WriteString(`": `)
		sb.WriteString(f.GoName)
		sb.WriteByte(',')
		sb.WriteByte('\n')
	}
	sb.WriteString("}\n\n")

	sb.WriteString("\n\nfunc (e ")
	sb.WriteString(t.GoName)
	sb.WriteString(") String() string {\n\treturn conv")
	sb.WriteString(t.GoName)
	sb.WriteString("ToString[e]\n}\n\n")
	sb.WriteString("func As")
	sb.WriteString(t.GoName)
	sb.WriteString("(s string) *")
	sb.WriteString(t.GoName)
	sb.WriteString(" {\n\tif v, ok := convStringTo")
	sb.WriteString(t.GoName)
	sb.WriteString("[s]; ok {\n\t\treturn &v\n\t}\n\treturn nil\n}\n\n")

	return sb.String()
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

func (t GoFile) GoString() string {
	sb := new(strings.Builder)

	sb.WriteString("package ")
	sb.WriteString(t.Package)
	sb.WriteString("\n\n")
	sb.WriteString("import (\n")
	for imp := range t.importsList {
		sb.WriteByte('\t')
		sb.WriteByte('"')
		sb.WriteString(imp)
		sb.WriteByte('"')
		sb.WriteByte('\n')
	}
	sb.WriteString(")\n\n")
	sb.WriteString("const minSize = ")
	sb.WriteString(strconv.Itoa(t.minAlloc))
	sb.WriteString("\n\n")
	for _, typ := range t.Types {
		sb.WriteString(typ.GoString())

		var placeholder = "x"
		sb.WriteString("\n\nfunc (")
		sb.WriteString(placeholder)
		sb.WriteByte(' ')
		sb.WriteString(typ.Name)
		sb.WriteString(") Bytes() []byte {\n\te := newEncoder(\n\t\tminSize +\n")
		for idx, field := range typ.Fields {
			var isStruct bool
			var isEnum bool
			if v, ok := t.UniqueTypes[field.GoType]; ok {
				if v {
					isEnum = true
				} else {
					isStruct = true
				}
			}

			if v, ok := t.UniqueTypes[field.ProtoName]; ok {
				if v {
					isEnum = true
				} else {
					isStruct = true
				}
			}
			switch field.GoType {
			case "bool":
				sb.WriteString("\t\t\t1")
				if idx == len(typ.Fields)-1 {
					sb.WriteString(")\n")
					continue
				}
				sb.WriteString(" +\n")

			case "string", "[]byte":
				sb.WriteString("\t\t\tbyteLen(len(")
				sb.WriteString(placeholder)
				sb.WriteByte('.')
				sb.WriteString(field.GoName)
				sb.WriteString(")) +\n")
				sb.WriteString("\t\t\tlen(")
				sb.WriteString(placeholder)
				sb.WriteByte('.')
				sb.WriteString(field.GoName)
				sb.WriteString(")")
				if idx == len(typ.Fields)-1 {
					sb.WriteString(")\n")
					continue
				}
				sb.WriteString(" +\n")

			case "int", "uint", "int32", "uint32", "int64", "uint64":
				sb.WriteString("\t\t\tbyteLen(")
				sb.WriteString(placeholder)
				sb.WriteByte('.')
				sb.WriteString(field.GoName)
				sb.WriteString(")")
				if idx == len(typ.Fields)-1 {
					sb.WriteString(")\n")
					continue
				}
				sb.WriteString(" +\n")
			default:
				if isEnum {
					sb.WriteString("\t\t\tbyteLen(")
					sb.WriteString(placeholder)
					sb.WriteByte('.')
					sb.WriteString(field.GoName)
					sb.WriteString(")")
					if idx == len(typ.Fields)-1 {
						sb.WriteString(")\n")
						continue
					}
					sb.WriteString(" +\n")
					continue
				}
				if isStruct {
					sb.WriteString("\t\t\tbyteLen(len(")
					sb.WriteString(placeholder)
					sb.WriteByte('.')
					sb.WriteString(field.GoName)
					sb.WriteString(".Bytes())) +\n")
					sb.WriteString("\t\t\tlen(")
					sb.WriteString(placeholder)
					sb.WriteByte('.')
					sb.WriteString(field.GoName)
					sb.WriteString(".Bytes())")
					if idx == len(typ.Fields)-1 {
						sb.WriteString(")\n")
						continue
					}
					sb.WriteString(" +\n")
				}
			}
		}
		for _, field := range typ.Fields {
			if field.IsRepeated {
				sb.WriteString("for idx := range ")
				sb.WriteString(placeholder)
				sb.WriteByte('.')
				sb.WriteString(field.GoName)
				sb.WriteString(" {\n\t")
			}
			if field.IsOptional {
				sb.WriteString("if ")
				sb.WriteString(placeholder)
				sb.WriteByte('.')
				sb.WriteString(field.GoName)
				if field.IsRepeated {
					sb.WriteString("[idx]")
				}
				sb.WriteString(" != nil {\n")
			}
			sb.WriteString("\te.b.WriteByte(")
			sb.WriteString(strconv.Itoa(field.idAndWire.Header()))
			sb.WriteString(")\n")
			if conc, ok := goTypes[field.GoType]; ok {
				sb.WriteString("\te.")
				sb.WriteString(conc.EncoderFunc())
				sb.WriteByte('(')
				if field.IsOptional {
					sb.WriteString("*")
				}
				sb.WriteString(placeholder)
				sb.WriteByte('.')
				sb.WriteString(field.GoName)
				if field.IsRepeated {
					sb.WriteString("[idx]")
				}
				sb.WriteByte(')')
				sb.WriteByte('\n')
			} else if isEnum, ok := t.UniqueTypes[field.GoType]; ok {
				if isEnum {
					sb.WriteString("\te.")
					sb.WriteString(Int64.EncoderFunc())
					sb.WriteByte('(')
					if field.IsOptional {
						sb.WriteString("*")
					}
					sb.WriteString(placeholder)
					sb.WriteByte('.')
					sb.WriteString(field.GoName)
					if field.IsRepeated {
						sb.WriteString("[idx]")
					}
					sb.WriteByte(')')
					sb.WriteByte('\n')
				} else {
					sb.WriteString("\te.")
					sb.WriteString(Bytes.EncoderFunc())
					sb.WriteByte('(')
					if field.IsOptional {
						sb.WriteString("*")
					}
					sb.WriteString(placeholder)
					sb.WriteByte('.')
					sb.WriteString(field.GoName)
					if field.IsRepeated {
						sb.WriteString("[idx]")
					}
					sb.WriteString(".Bytes())")
					sb.WriteByte('\n')
				}
			}

			if field.IsOptional {
				sb.WriteString("}\n")
			}

			if field.IsRepeated {
				sb.WriteString("}\n")
			}
		}
		sb.WriteString("\n\treturn e.b.Bytes()\n}\n")

	}
	for _, enum := range t.Enums {
		sb.WriteString(enum.GoString())
	}

	sb.WriteString("\n\ntype encoder struct {\n\tb *bytes.Buffer\n}")
	sb.WriteString("\n\nfunc newEncoder(size int) encoder {\n\tif size == 0 {\n\t\tsize = minSize\n\t}\n\treturn encoder{\n\t\tb: bytes.NewBuffer(make([]byte, 0, size)),\n\t}\n}")
	sb.WriteString("\n\ntype signed interface {\n\t~int | ~int16 | ~int32 | ~int64\n}")
	sb.WriteString("\n\ntype unsigned interface {\n\t~uint | ~uint16 | ~uint32 | ~uint64\n}")
	sb.WriteString("\n\ntype number interface {\n\tsigned | unsigned\n}")
	sb.WriteString("\n\nfunc byteLen[T number](v T) (size int) {\n\tfor i := 0 ; i < 8 ; i++ {\n\t\tv = v >> 8\n\t\tif v == 0 {\n\t\t\treturn i + i\n\t\t}\n\t}\n\treturn 0\n}")

	return sb.String()
}
