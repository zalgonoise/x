package protofile

import (
	"fmt"
	"io"
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

func (t GoType) TypeGoString() string {
	sb := new(strings.Builder)

	sb.WriteString(fmt.Sprintf(`

// %s describes the message
type %s struct {
`, t.Name, t.Name))

	for _, f := range t.Fields {
		sb.WriteString(fmt.Sprintf(
			`	%s	`, f.GoName))

		if f.IsRepeated {
			sb.WriteString("[]")
		}
		if f.IsOptional {
			sb.WriteString("*")
		}
		sb.WriteString(fmt.Sprintf(
			`%s	// id: %d; wire_type: %d
`, f.GoType, f.ProtoIndex, f.idAndWire.Wire))
	}
	sb.WriteString("}\n")

	return sb.String()
}

func nextPlusSeparator(sb io.StringWriter, idx, len int) {
	if idx == len-1 {
		_, _ = sb.WriteString(")\n")
		return
	}
	_, _ = sb.WriteString(" +\n")
}

func (g *GoType) EncoderGoString(f GoFile) string {
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintf(
		`

func (x %s) Bytes() []byte {
e := newEncoder(
	minSize +
`, g.Name))
	for idx, field := range g.Fields {
		var isStruct bool
		var isEnum bool
		if v, ok := f.UniqueTypes[field.GoType]; ok {
			if v {
				isEnum = true
			} else {
				isStruct = true
			}
		}
		if field.IsRepeated {

			sb.WriteString(fmt.Sprintf(
				`			len(x.%s) * 8`, field.GoName))
			nextPlusSeparator(sb, idx, len(g.Fields))
			continue
		}
		if field.IsOptional {
			sb.WriteString("\t\t\t8")
			nextPlusSeparator(sb, idx, len(g.Fields))
			continue
		}
		if isEnum {
			sb.WriteString(fmt.Sprintf(
				`			byteLen(uint64(x.%s))`, field.GoName))
			nextPlusSeparator(sb, idx, len(g.Fields))
			continue
		}
		if isStruct {
			sb.WriteString(fmt.Sprintf(
				`			byteLen(len(x.%s.Bytes())) +
		len(x.%s.Bytes())`, field.GoName, field.GoName))
			nextPlusSeparator(sb, idx, len(g.Fields))
		}

		switch field.GoType {
		case "bool":
			sb.WriteString("\t\t\t1")
			nextPlusSeparator(sb, idx, len(g.Fields))

		case "string", "[]byte":
			sb.WriteString(fmt.Sprintf(
				`			byteLen(len(x.%s)) +
		len(x.%s)`, field.GoName, field.GoName))
			nextPlusSeparator(sb, idx, len(g.Fields))

		case "int", "uint", "int32", "uint32", "int64", "uint64":
			sb.WriteString(fmt.Sprintf(
				`			byteLen(x.%s)`, field.GoName))
			nextPlusSeparator(sb, idx, len(g.Fields))
		case "float32":
			sb.WriteString("\t\t\t4")
			nextPlusSeparator(sb, idx, len(g.Fields))
		default:
			sb.WriteString("\t\t\t8")
			nextPlusSeparator(sb, idx, len(g.Fields))
		}
	}
	for _, field := range g.Fields {
		if field.IsRepeated {
			sb.WriteString(fmt.Sprintf(
				`for idx := range x.%s {
`, field.GoName))
		}
		if field.IsOptional {
			sb.WriteString(fmt.Sprintf(
				`if x.%s`, field.GoName))

			if field.IsRepeated {
				sb.WriteString("[idx]")
			}
			sb.WriteString(" != nil {\n")
		}
		var forceCloseIf bool
		switch field.GoType {
		case "bool":
			sb.WriteString(fmt.Sprintf(
				`if x.%s`, field.GoName))
			if field.IsRepeated {
				sb.WriteString("[idx]")
			}
			sb.WriteString(" {\n\t")
		case "string":
			sb.WriteString(fmt.Sprintf(
				`if x.%s`, field.GoName))
			if field.IsRepeated {
				sb.WriteString("[idx]")
			}
			sb.WriteString(` != "" {
`)
		case "[]byte":
			sb.WriteString(fmt.Sprintf(
				`if len(x.%s`, field.GoName))
			if field.IsRepeated {
				sb.WriteString("[idx]")
			}
			sb.WriteString(`) > 0 {
`)
		case "int", "uint", "int32", "uint32", "int64", "uint64", "float32", "float64":
			sb.WriteString(fmt.Sprintf(
				`if x.%s`, field.GoName))

			if field.IsRepeated {
				sb.WriteString("[idx]")
			}
			sb.WriteString(` != 0 {
`)
		default:
			// enums and structs
			if isEnum, ok := f.UniqueTypes[field.GoType]; ok {
				if !isEnum {
					sb.WriteString(fmt.Sprintf(
						`	if structBytes := x.%s`, field.GoName))
					if field.IsRepeated {
						sb.WriteString("[idx]")
					}
					sb.WriteString(`.Bytes(); len(structBytes) > 0 {
`)
					forceCloseIf = true
				}
			}
		}

		sb.WriteString(fmt.Sprintf(
			`	e.b.WriteByte(%d)
`, field.idAndWire.Header()))
		if conc, ok := goTypes[field.GoType]; ok {
			sb.WriteString(fmt.Sprintf(
				`	e.%s(`, conc.EncoderFunc()))
			if field.IsOptional {
				sb.WriteString("*")
			}
			sb.WriteString(fmt.Sprintf("x.%s", field.GoName))
			if field.IsRepeated {
				sb.WriteString("[idx]")
			}
			sb.WriteString(`)
`)
		} else if isEnum, ok := f.UniqueTypes[field.GoType]; ok {
			if isEnum {
				sb.WriteString(fmt.Sprintf(
					`	e.%s(uint64(`, Uint64.EncoderFunc()))
				if field.IsOptional {
					sb.WriteString("*")
				}
				sb.WriteString(fmt.Sprintf("x.%s", field.GoName))
				if field.IsRepeated {
					sb.WriteString("[idx]")
				}
				sb.WriteString(`))
`)
			} else {
				sb.WriteString(fmt.Sprintf(
					`	e.%s(structBytes)
`, Bytes.EncoderFunc()))
			}
		}
		switch field.GoType {
		case "bool", "string", "[]byte", "int", "uint", "int32", "uint32", "int64", "uint64", "float32", "float64":
			sb.WriteString("}\n")
		default:
			if forceCloseIf {
				sb.WriteString("}\n")
			}
		}

		if field.IsOptional {
			sb.WriteString("}\n")
		}

		if field.IsRepeated {
			sb.WriteString("}\n")
		}
	}
	sb.WriteString(`
	return e.b.Bytes()
}
`)

	return sb.String()
}
func (g *GoType) DecoderGoString(f GoFile) string {
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintf(`

	type dec%s struct {
		*bytes.Reader
	}
	
	func To%s(buf []byte) (%s, error) {
		return (&dec%s{Reader: bytes.NewReader(buf)}).decode()
	}
	
	func (d *dec%s) decode() (%s, error) {
		x, err := decode%s(d.Reader)
		if err != nil && !errors.Is(err, io.EOF) {
			return x, err
		}
		return x, nil
	}
	
	func decode%s(r io.ByteReader) (%s, error) {
		var x = %s{}
		for {
			v, err := r.ReadByte()
			if err != nil {
				return x, err
			}
			switch v {
	`, g.Name, g.Name, g.Name, g.Name, g.Name, g.Name, g.Name, g.Name, g.Name, g.Name))

	for _, field := range g.Fields {
		var decoderFn string
		var isEnum bool
		var isStruct bool
		if conc, ok := goTypes[field.GoType]; ok {
			decoderFn = conc.DecoderFunc()
		} else if enum, ok := f.UniqueTypes[field.GoType]; ok {
			if enum {
				isEnum = true
				decoderFn = Uint64.DecoderFunc()
			} else {
				isStruct = true
				decoderFn = "decode" + field.GoType
			}
		}
		sb.WriteString(fmt.Sprintf(
			`		case header%s%s:
		`, g.Name, field.GoName))
		if isStruct {
			sb.WriteString(`_, err := r.ReadByte() // length byte
		if err != nil {
			return x, err
		}
		`)
		}
		sb.WriteString(fmt.Sprintf(
			`%s, err := %s(r`, field.ProtoName, decoderFn))
		switch field.GoType {
		case "[]byte", "string":
			sb.WriteString(".(io.Reader)")
		}
		sb.WriteString(fmt.Sprintf(
			`)
		if err != nil && !errors.Is(err, io.EOF) {
			return x, err
		}
		x.%s = `, field.GoName))
		if field.IsRepeated {
			sb.WriteString(fmt.Sprintf(`append(x.%s, `, field.GoName))
		}
		if isEnum {
			sb.WriteString("(")
			if field.IsOptional {
				sb.WriteString("*")
			}
			sb.WriteString(fmt.Sprintf(`%s)(`, field.GoType))
		}
		if field.IsOptional {
			sb.WriteString("&")
		}
		sb.WriteString(field.ProtoName)
		if isEnum {
			sb.WriteString(")")
		}
		if field.IsRepeated {
			sb.WriteString(")")
		}
		sb.WriteByte('\n')

	}

	sb.WriteString(`		case 0:
		return x, nil
	default:
		return x, errors.New("invalid header")
	}
}
}`)

	return sb.String()
}
