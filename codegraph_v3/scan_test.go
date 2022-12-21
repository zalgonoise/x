package codegraph

import (
	"testing"

	cur "github.com/zalgonoise/cur"
	"github.com/zalgonoise/x/ptr"
)

var (
// //go:embed testing/testdata.go
// goFile []byte
)

const (
	// path = "./testing/testdata.go"
	// path = "./testing/testdata/testdata_short.go"
	path = "./testing/testdata/testdata_generic.go"
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

func TestExplore(t *testing.T) {
	code := `package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, World!")
}`

	_, err := Explore([]byte(code))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParams(t *testing.T) {
	paramSet := [][]byte{
		[]byte(`(int)`),
		[]byte(`(*int)`),
		[]byte(`(context.Context)`),
		[]byte(`(ctx context.Context)`),
		[]byte(`(n, len int)`),
		[]byte(`(x []items)`),
		[]byte(`(x []*items)`),
		[]byte(`(procs ...Processor[T])`),
		[]byte(`(json map[string]any)`),
		[]byte(`(json map[string]interface{})`),
		// []byte(`(testFn func(s string) (bool, error))`),
	}
	t.Run("int", func(t *testing.T) {
		wants := &Type{
			Type: "int",
		}

		tokens, err := Explore(paramSet[0])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParams(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}

	})

	t.Run("*int", func(t *testing.T) {
		wants := &Type{
			Type:      "int",
			IsPointer: ptr.To(true),
		}

		tokens, err := Explore(paramSet[1])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParams(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}
		if types[0].IsPointer == nil {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.IsPointer, types[0].IsPointer)
			return
		}
		if *types[0].IsPointer != *wants.IsPointer {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.IsPointer, types[0].IsPointer)
		}

	})

	t.Run("context.Context", func(t *testing.T) {
		wants := &Type{
			Type:    "Context",
			Package: ptr.To("context"),
		}

		tokens, err := Explore(paramSet[2])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParams(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}
		if types[0].Package == nil {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Package, types[0].Package)
			return
		}
		if *types[0].Package != *wants.Package {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Package, types[0].Package)
		}

	})

	t.Run("ctx context.Context", func(t *testing.T) {
		wants := &Type{
			Name:    ptr.To("ctx"),
			Type:    "Context",
			Package: ptr.To("context"),
		}

		tokens, err := Explore(paramSet[3])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParams(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}
		if types[0].Package == nil {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Package, types[0].Package)
			return
		}
		if *types[0].Package != *wants.Package {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Package, types[0].Package)
		}
		if types[0].Name == nil {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Name, types[0].Name)
			return
		}
		if *types[0].Name != *wants.Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Name, types[0].Name)
		}

	})
}
