package protofile

import (
	_ "embed"
	"testing"
)

//go:embed testdata/all.proto
var protofile []byte

func TestParser(t *testing.T) {
	str, err := Run(protofile)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(str)
	t.Error()
}
