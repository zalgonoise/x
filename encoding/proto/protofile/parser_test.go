package protofile

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/zalgonoise/gio"
)

//go:embed testdata/all.proto
var protofile []byte

func TestParser(t *testing.T) {
	r := (gio.Reader[byte])(bytes.NewReader(protofile))
	n, err := Parse(r)
	if err != nil {
		t.Error(err)
		return
	}
	if n == 0 {
		t.Errorf("zero bytes written")
	}
}
