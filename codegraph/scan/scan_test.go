package scan

import (
	_ "embed"
	"go/token"
	"testing"
)

var (
	//   //go:embed testdata/gofiles/simple.go
	// goFile []byte

	filePath string = "testdata/gofiles/simple.go"
)

func TestReadFile(t *testing.T) {
	var testPrinterFunc ParseFunc = func(pos token.Pos, tok token.Token, lit string) {
		t.Logf("%v: '%s' -> %s\n", pos, tok, lit)
	}

	procF, err := New(filePath)
	if err != nil {
		t.Errorf("unexpected error reading file from path %s: %v", filePath, err)
	}

	err = procF.Parse(testPrinterFunc)
	if err != nil {
		t.Errorf("unexpected error parsing Go file: %v", err)
	}

	t.Log(procF)
	t.Error()
}
