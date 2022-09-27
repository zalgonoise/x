package scan

import (
	_ "embed"
	"testing"
)

var (
	//   //go:embed testdata/gofiles/simple.go
	// goFile []byte

	filePath string = "testdata/gofiles/simple.go"
)

func TestReadFile(t *testing.T) {

	procF, err := New(filePath)
	if err != nil {
		t.Errorf("unexpected error reading file from path %s: %v", filePath, err)
	}

	err = procF.Parse()
	if err != nil {
		t.Errorf("unexpected error parsing Go file: %v", err)
	}

	t.Log(procF)
	t.Error()
}
