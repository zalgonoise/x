package protofile

import (
	"strconv"
	"strings"
)

type IDAndWire struct {
	ID   int
	Wire int
	Name string
}

func (iaw IDAndWire) Header() int {
	header := (iaw.ID << 3) | iaw.Wire
	if header > 255 {
		return 0
	}
	return header
}

func fmtPascal(name string) string {
	split := strings.Split(name, "_")
	for idx, s := range split {
		n := []byte(s)
		if n[0] > 96 {
			n[0] = n[0] - 32 // fast uppercase
		}
		split[idx] = string(n)
	}
	return strings.Join(split, "")
}

func HeaderGoString(fields ...IDAndWire) string {
	if len(fields) == 0 {
		return ""
	}

	sb := new(strings.Builder)
	sb.WriteString("var (\n")

	for _, f := range fields {
		if f.Name == "" {
			continue
		}
		n := fmtPascal(f.Name)
		b := (f.ID << 3) | f.Wire
		sb.WriteString("\theader")
		sb.WriteString(string(n))
		sb.WriteString(" uint64 = ")
		sb.WriteString(strconv.Itoa(b))
		sb.WriteByte('\n')
	}
	sb.WriteString(")\n\n")

	return sb.String()
}
