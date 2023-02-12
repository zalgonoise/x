package protofile

import (
	"bytes"
	_ "embed"
	"os"
	"testing"

	"github.com/zalgonoise/gio"
)

//go:embed testdata/all.proto
var protofile []byte

func TestParser(t *testing.T) {
	r := (gio.Reader[byte])(bytes.NewReader(protofile))
	buf, err := Parse[ProtoToken, byte, gio.Reader[byte]](r)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = gio.Copy[byte](os.Stderr, buf)
	if err != nil {
		t.Error(err)
		return
	}

	t.Error()
}
