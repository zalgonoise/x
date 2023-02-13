package protofile

import (
	"fmt"
	"strings"
)

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

func (t GoEnum) TypeGoString() string {
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintf(`

// %s outlines the enumeration
type %s uint64


const (
`, t.GoName, t.GoName))

	for _, f := range t.Fields {
		sb.WriteString(fmt.Sprintf(
			`	%s	%s	=	%d
`, f.GoName, t.GoName, f.Index))
	}
	sb.WriteString(fmt.Sprintf(
		`)

var (
	conv%sToString = map[%s]string{
`, t.GoName, t.GoName))

	for _, f := range t.Fields {
		sb.WriteString(fmt.Sprintf(
			`	%s: "%s",
`, f.GoName, f.GoName))
	}
	sb.WriteString(fmt.Sprintf(
		`}

	convStringTo%s = map[string]%s{
`, t.GoName, t.GoName))

	for _, f := range t.Fields {
		sb.WriteString(fmt.Sprintf(
			`	"%s": %s,
`, f.GoName, f.GoName))
	}
	sb.WriteString(fmt.Sprintf(
		`}
)

func (e %s) String() string {
	return conv%sToString[e]
}

func As%s(s string) *%s {
	if v, ok := convStringTo%s[s]; ok {
		return &v
	}
	return nil
}

`, t.GoName, t.GoName, t.GoName, t.GoName, t.GoName))

	return sb.String()
}
