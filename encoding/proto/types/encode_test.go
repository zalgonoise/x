package types

import (
	"strings"
	"testing"
)

func printBin(t *testing.T, data []byte) {
	s := new(strings.Builder)
	s.WriteString("\n")
	args := []any{}

	var i int
	for _, b := range data {
		if i > 3 {
			s.WriteString("\n")
			i = 0
		}
		s.WriteString("%08b\t")
		i++
		args = append(args, b)
	}
	t.Logf(s.String(), args...)
}

func TestEncode(t *testing.T) {
	b := NewEncoder()

	b.EncodeVarintField(5, 103)
	b.EncodeField(2, 2, []byte("pb by hand"))
	b.EncodeVarintField(3, 301)
	b.EncodeVarintField(4, 1)

	t.Log(b.String())
	buf := b.Bytes()
	printBin(t, buf)

	d := NewDecoder(buf)
	out, err := d.Decode()
	if err != nil {
		t.Error(err)
	}

	if len(out) == 0 {
		t.Error("EMPTY MAP")
		return
	}
	f5 := &Field[uint64]{}
	f2 := &Field[[]byte]{}
	err = (*(out[5]).(*GField)).To(f5)
	if err != nil {
		t.Error(err)
	}
	err = (*(out[2]).(*GField)).To(f2)
	if err != nil {
		t.Error(err)
	}

	t.Log(out, f5, f5.Value, f5.Num)
	t.Log(out, f2, string(f2.Value), f2.Num)

	t.Error()
}
