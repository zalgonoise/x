package codegraph

import (
	"testing"
)

var (
// //go:embed testing/testdata.go
// goFile []byte
)

const (
	// path = "./testing/testdata.go"
	path = "./testing/testdata/testdata_short.go"
)

func TestExtract(t *testing.T) {
	tok, err := Extract(path)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	for _, token := range tok {
		t.Log(token.Pos, token.Tok.String(), token.Lit)
	}
	t.Error()
}

func TestGetPackage(t *testing.T) {
	wt := New(path)
	err := wt.Package()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	t.Log(wt.GoFile.String())
	t.Error()
}

func TestGetPackageAndFuncInput(t *testing.T) {
	wt := New(path)
	err := wt.Package()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err = wt.Func()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	t.Log(wt.GoFile.String())
	t.Error()
}
