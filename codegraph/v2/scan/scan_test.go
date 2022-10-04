package scan

import (
	_ "embed"
	"go/token"
	"testing"

	scan "github.com/zalgonoise/x/codegraph/v2/scan/extractors"
)

var (
	//   //go:embed testdata/gofiles/simple.go
	// goFile []byte

	filePath string = "testdata/gofiles/simple.go"
)

func TestReadFile(t *testing.T) {
	var testPrinterFunc scan.ParseFunc = func(pos token.Pos, tok token.Token, lit string) {
		// t.Logf("%v: '%s' -> %s\n", pos, tok, lit)
	}

	procF, err := scan.New(filePath)
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
