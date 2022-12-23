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
		[]byte(`(n [max]int)`),
		[]byte(`(n *[3]int)`),
		[]byte(`(*[]byte)`),
		[]byte(`(*map[any]any)`),
		[]byte(`(a, b, c int, data []byte)`),
		[]byte(`()`),
		[]byte(`(A, B, C, D)`), // span of generic types
		[]byte(`(testFn func(s string) (bool, error))`),
		[]byte(`(testFn func(s string) error)`),
		[]byte(`(testFn func(s string) func(int) error)`),
		// []byte(`(json map[string]interface{})`),
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

		types := ExtractParamsReverse(c)
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

		types := ExtractParamsReverse(c)
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

		types := ExtractParamsReverse(c)
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
			Name:    "ctx",
			Type:    "Context",
			Package: ptr.To("context"),
		}

		tokens, err := Explore(paramSet[3])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParamsReverse(c)
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
		if types[0].Name == "" {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Name, types[0].Name)
			return
		}
		if types[0].Name != wants.Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Name, types[0].Name)
		}
	})

	t.Run("n, len int", func(t *testing.T) {
		wants := []*Type{
			{
				Type: "int",
				Name: "n",
			},
			{
				Type: "int",
				Name: "len",
			},
		}

		tokens, err := Explore(paramSet[4])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParamsReverse(c)
		if len(types) != 2 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants[0].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[0], types[0])
		}
		if types[0].Name == "" {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[0].Name, types[0].Name)
			return
		}
		if types[0].Name != wants[0].Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[0].Name, types[0].Name)
		}
		if types[1].Type != wants[1].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[1], types[1])
		}
		if types[1].Name == "" {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[1].Name, types[1].Name)
			return
		}
		if types[1].Name != wants[1].Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[1].Name, types[1].Name)
		}
	})

	t.Run("x []items", func(t *testing.T) {
		wants := &Type{
			Name:  "x",
			Type:  "items",
			Slice: &RSlice{},
		}

		tokens, err := Explore(paramSet[5])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParamsReverse(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}

		if types[0].Slice == nil {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Slice, types[0].Slice)
			return
		}
		if types[0].Name == "" {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Name, types[0].Name)
			return
		}
		if types[0].Name != wants.Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Name, types[0].Name)
		}
	})

	t.Run("x []*items", func(t *testing.T) {
		wants := &Type{
			Name:      "x",
			IsPointer: ptr.To(true),
			Type:      "items",
			Slice:     &RSlice{},
		}

		tokens, err := Explore(paramSet[6])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParamsReverse(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}

		if types[0].Slice == nil {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Slice, types[0].Slice)
			return
		}
		if types[0].IsPointer == nil || !*types[0].IsPointer {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants.IsPointer, types[0].IsPointer)
			return
		}
		if types[0].Name == "" || types[0].Name != wants.Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Name, types[0].Name)
			return
		}
	})

	t.Run("procs ...Processor[T]", func(t *testing.T) {
		wants := &Type{
			Name: "procs",
			Type: "Processor",
			Slice: &RSlice{
				IsVariadic: ptr.To(true),
			},
			Generics: []*Type{
				{Type: "T"},
			},
		}

		tokens, err := Explore(paramSet[7])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParamsReverse(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}

		if types[0].Slice == nil || !*types[0].Slice.IsVariadic {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Slice, types[0].Slice)
			return
		}
		if types[0].Generics == nil || len(types[0].Generics) != 1 || types[0].Generics[0].Type != wants.Generics[0].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Generics, types[0].Generics)
			return
		}

		if types[0].Name == "" || types[0].Name != wants.Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Name, types[0].Name)
			return
		}
	})

	t.Run("json map[string]any", func(t *testing.T) {
		wants := &Type{
			Name: "json",
			Type: "map",
			Map: &RMap{
				Key: "string",
				Value: Type{
					Type: "any",
				},
			},
		}

		tokens, err := Explore(paramSet[8])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParamsReverse(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}

		if types[0].Name == "" || types[0].Name != wants.Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Name, types[0].Name)
			return
		}

		if types[0].Map == nil || types[0].Map.Key != wants.Map.Key {
			t.Errorf("unexpected output error: wanted %v ; got %v", types[0].Map, wants.Map.Key)
			return
		}
		if types[0].Map == nil || types[0].Map.Value.Type != wants.Map.Value.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", types[0].Map, wants.Map.Value.Type)
			return
		}

	})

	t.Run("n [max]int", func(t *testing.T) {
		wants := &Type{
			Name: "n",
			Type: "int",
			Slice: &RSlice{
				LenName: ptr.To("max"),
			},
		}

		tokens, err := Explore(paramSet[9])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParamsReverse(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}

		if types[0].Slice == nil || types[0].Slice.LenName == nil {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Slice.LenName, types[0].Slice)
			return
		}
		if *types[0].Slice.LenName != *wants.Slice.LenName {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Slice.LenName, types[0].Slice)
			return
		}
		if types[0].Name == "" || types[0].Name != wants.Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Name, types[0].Name)
			return
		}
	})
	t.Run("n *[3]int", func(t *testing.T) {
		wants := &Type{
			Name: "n",
			Type: "int",
			Slice: &RSlice{
				Len:       ptr.To(3),
				IsPointer: ptr.To(true),
			},
		}

		tokens, err := Explore(paramSet[10])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParamsReverse(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}

		if types[0].Slice == nil || types[0].Slice.Len == nil || types[0].Slice.IsPointer == nil {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Slice.Len, types[0].Slice)
			return
		}
		if *types[0].Slice.Len != *wants.Slice.Len || !*types[0].Slice.IsPointer {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Slice.LenName, types[0].Slice)
			return
		}
		if types[0].Name == "" || types[0].Name != wants.Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Name, types[0].Name)
			return
		}
	})

	t.Run("*[]byte", func(t *testing.T) {
		wants := &Type{
			Type: "byte",
			Slice: &RSlice{
				IsPointer: ptr.To(true),
			},
		}

		tokens, err := Explore(paramSet[11])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParamsReverse(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}

		if types[0].Slice == nil || types[0].Slice.IsPointer == nil {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Slice.Len, types[0].Slice)
			return
		}
		if !*types[0].Slice.IsPointer {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants.Slice.LenName, types[0].Slice)
			return
		}
	})

	t.Run("*map[any]any", func(t *testing.T) {
		wants := &Type{
			Type: "map",
			Map: &RMap{
				IsPointer: ptr.To(true),
				Key:       "any",
				Value: Type{
					Type: "any",
				},
			},
		}

		tokens, err := Explore(paramSet[12])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParamsReverse(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, types[0])
		}

		if types[0].Map == nil || types[0].Map.Key != wants.Map.Key {
			t.Errorf("unexpected output error: wanted %v ; got %v", types[0].Map, wants.Map.Key)
			return
		}
		if types[0].Map == nil || types[0].Map.Value.Type != wants.Map.Value.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", types[0].Map, wants.Map.Value.Type)
			return
		}
		if types[0].Map.IsPointer == nil || *types[0].Map.IsPointer != *wants.Map.IsPointer {
			t.Errorf("unexpected output error: wanted %v ; got %v", types[0].Map, wants.Map.IsPointer)
			return
		}

	})

	t.Run("a, b, c int, data []byte", func(t *testing.T) {
		wants := []*Type{
			{
				Name: "a",
				Type: "int",
			},
			{
				Name: "b",
				Type: "int",
			},
			{
				Name: "c",
				Type: "int",
			},
			{
				Name:  "data",
				Type:  "byte",
				Slice: &RSlice{},
			},
		}

		tokens, err := Explore(paramSet[13])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		types := ExtractParamsReverse(c)
		if len(types) != 4 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		if types[0].Type != wants[0].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[0], types[0])
		}
		if types[0].Name == "" || types[0].Name != wants[0].Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[0].Name, types[0].Name)
		}

		if types[1].Type != wants[1].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[1], types[1])
		}
		if types[1].Name == "" || types[1].Name != wants[1].Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[1], types[1])
		}

		if types[2].Type != wants[2].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[2], types[2])
		}
		if types[2].Name == "" || types[2].Name != wants[2].Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[2], types[2])
		}

		if types[3].Type != wants[3].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[3], types[3])
		}
		if types[3].Name == "" || types[3].Name != wants[3].Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[3], types[3])
		}
		if types[3].Slice == nil {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants[3].Slice, types[3])
		}
	})

	t.Run("nil", func(t *testing.T) {
		tokens, err := Explore(paramSet[14])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		t.Log(tokens)
		types := ExtractParamsReverse(c)
		if types != nil {
			t.Errorf("expected nil output; got %v", types)
		}
	})

	t.Run("(A, B, C, D)", func(t *testing.T) {
		wants := []*Type{
			{
				Type: "A",
			},
			{
				Type: "B",
			},
			{
				Type: "C",
			},
			{
				Type: "D",
			},
		}
		tokens, err := Explore(paramSet[15])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)
		types := ExtractParamsReverse(c)
		if len(types) != 4 {
			t.Errorf("unexpected length: %d", len(types))
			return
		}

		for idx, w := range wants {
			if types[idx].Type != w.Type {
				t.Errorf("unexpected type on idx %v: wanted %v ; got %v", idx, w.Type, *types[idx])
			}
		}
	})

	t.Run("(testFn func(s string) (bool, error))", func(t *testing.T) {
		wants := &Type{
			Name: "testFn",
			Type: "func",
			Func: &RFunc{
				IsFunc: ptr.To(true),
				InputParams: []*Type{
					{
						Name: "s",
						Type: "string",
					},
				},
				Returns: []*Type{
					{
						Type: "bool",
					},
					{
						Type: "error",
					},
				},
			},
		}
		tokens, err := Explore(paramSet[16])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)
		types := ExtractParamsReverse(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			for _, tp := range types {
				t.Log(*tp)
			}
			return
		}

		if types[0].Name != wants.Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func == nil {
			t.Errorf("unexpected nil function element: wanted %v ; got %v", *wants, *types[0])
		}
		if len(types[0].Func.InputParams) != len(wants.Func.InputParams) {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.InputParams[0].Name != wants.Func.InputParams[0].Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.InputParams[0].Type != wants.Func.InputParams[0].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if len(types[0].Func.Returns) != len(wants.Func.Returns) {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.Returns[0].Type != wants.Func.Returns[0].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.Returns[1].Type != wants.Func.Returns[1].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
	})

	t.Run("(testFn func(s string) error)", func(t *testing.T) {
		wants := &Type{
			Name: "testFn",
			Type: "func",
			Func: &RFunc{
				IsFunc: ptr.To(true),
				InputParams: []*Type{
					{
						Name: "s",
						Type: "string",
					},
				},
				Returns: []*Type{
					{
						Type: "error",
					},
				},
			},
		}
		tokens, err := Explore(paramSet[17])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)
		types := ExtractParamsReverse(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			for _, tp := range types {
				t.Log(*tp)
			}
			return
		}

		if types[0].Name != wants.Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func == nil {
			t.Errorf("unexpected nil function element: wanted %v ; got %v", *wants, *types[0])
		}
		if len(types[0].Func.InputParams) != len(wants.Func.InputParams) {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.InputParams[0].Name != wants.Func.InputParams[0].Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.InputParams[0].Type != wants.Func.InputParams[0].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if len(types[0].Func.Returns) != len(wants.Func.Returns) {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.Returns[0].Type != wants.Func.Returns[0].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
	})

	t.Run("(testFn func(s string) func(int) error)", func(t *testing.T) {
		wants := &Type{
			Name: "testFn",
			Type: "func",
			Func: &RFunc{
				IsFunc: ptr.To(true),
				InputParams: []*Type{
					{
						Name: "s",
						Type: "string",
					},
				},
				Returns: []*Type{
					{
						Type: "func",
						Func: &RFunc{
							IsFunc: ptr.To(true),
							InputParams: []*Type{
								{
									Type: "int",
								},
							},
							Returns: []*Type{
								{
									Type: "error",
								},
							},
						},
					},
				},
			},
		}
		tokens, err := Explore(paramSet[18])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)
		types := ExtractParamsReverse(c)
		if len(types) != 1 {
			t.Errorf("unexpected length: %d", len(types))
			for _, tp := range types {
				t.Log(*tp)
			}
			return
		}

		if types[0].Name != wants.Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Type != wants.Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func == nil {
			t.Errorf("unexpected nil function element: wanted %v ; got %v", *wants, *types[0])
		}
		if len(types[0].Func.InputParams) != len(wants.Func.InputParams) {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.InputParams[0].Name != wants.Func.InputParams[0].Name {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.InputParams[0].Type != wants.Func.InputParams[0].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if len(types[0].Func.Returns) != len(wants.Func.Returns) {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.Returns[0].Type != wants.Func.Returns[0].Type {
			t.Errorf("unexpected output error: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.Returns[0].Func == nil {
			t.Errorf("unexpected nil return function element: wanted %v ; got %v", *wants, *types[0])
		}
		if len(types[0].Func.Returns[0].Func.InputParams) != len(wants.Func.Returns[0].Func.InputParams) {
			t.Errorf("unexpected nil return function element: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.Returns[0].Func.InputParams[0].Type != wants.Func.Returns[0].Func.InputParams[0].Type {
			t.Errorf("unexpected nil return function element: wanted %v ; got %v", *wants, *types[0])
		}
		if len(types[0].Func.Returns[0].Func.Returns) != len(wants.Func.Returns[0].Func.Returns) {
			t.Errorf("unexpected nil return function element: wanted %v ; got %v", *wants, *types[0])
		}
		if types[0].Func.Returns[0].Func.Returns[0].Type != wants.Func.Returns[0].Func.Returns[0].Type {
			t.Errorf("unexpected nil return function element: wanted %v ; got %v", *wants, *types[0])
		}
	})
}

func TestFuncs(t *testing.T) {
	fnSet := [][]byte{
		[]byte(`func Add(x, y int) int { return x + y }`),
	}

	t.Run("Simple", func(t *testing.T) {
		wants := &Type{
			Name: "Add",
			Kind: TypeFunction,
			Func: &RFunc{
				InputParams: []*Type{
					{
						Name: "x",
						Type: "int",
					},
					{
						Name: "y",
						Type: "int",
					},
				},
				Returns: []*Type{
					{
						Type: "int",
					},
				},
			},
		}
		tokens, err := Explore(fnSet[0])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		c := cur.NewCursor(tokens)

		fnType := ExtractFunc(c)

		if fnType == nil {
			t.Errorf("expected type not to be nil")
			return
		}

		if wants.Name != fnType.Name {
			t.Errorf("unexpected name: wanted %v ; got %v", *wants, *fnType)
		}
		if wants.Kind != fnType.Kind {
			t.Errorf("unexpected kind: wanted %v ; got %v", *wants, *fnType)
		}
		if fnType.Func == nil {
			t.Errorf("unexpected nil Func element: wanted %v ; got %v", *wants, *fnType)
		}
		if len(fnType.Func.InputParams) != 2 {
			t.Errorf("unexpected input param length: wanted %v ; got %v", 2, len(fnType.Func.InputParams))
			return
		}
		if wants.Func.InputParams[0].Name != fnType.Func.InputParams[0].Name ||
			wants.Func.InputParams[0].Type != fnType.Func.InputParams[0].Type {
			t.Errorf("unexpected input param: wanted %v ; got %v", *wants.Func.InputParams[0], *fnType.Func.InputParams[0])
		}
		if wants.Func.InputParams[1].Name != fnType.Func.InputParams[1].Name ||
			wants.Func.InputParams[1].Type != fnType.Func.InputParams[1].Type {
			t.Errorf("unexpected input param: wanted %v ; got %v", *wants.Func.InputParams[1], *fnType.Func.InputParams[1])
		}
		if len(fnType.Func.Returns) != 1 {
			t.Errorf("unexpected returns length: wanted %v ; got %v", 1, len(fnType.Func.Returns))
			return
		}
		if wants.Func.Returns[0].Type != fnType.Func.Returns[0].Type {
			t.Errorf("unexpected return element: wanted %v ; got %v", *wants.Func.Returns[0], *fnType.Func.Returns[0])
		}
	})
}

func TestFuncAsParam(t *testing.T) {
	tokens, err := Explore([]byte(`func Sort(a, b int, func(a, b int) bool) bool`))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	for _, tok := range tokens {
		t.Log(tok.Tok.String(), tok.Lit)
	}
	t.Error()
}

func TestInterface(t *testing.T) {
	tokens, err := Explore([]byte(`type ReadWriter interface {
	Read([]byte) func(byte) error
	Write([]byte) (int, error)
}`))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	for _, tok := range tokens {
		t.Log(tok.Tok.String(), tok.Lit)
	}
	t.Error()
}

func TestStruct(t *testing.T) {
	wants := &Type{
		Name: "Person",
		Kind: TypeStruct,
		Struct: &RStruct{
			IsStruct: ptr.To(true),
			Elems: []*Type{
				{
					Name: "Name",
					Type: "string",
				},
				{
					Name: "Age",
					Type: "int",
				},
				{
					Type: "Job",
				},
				{
					Name: "action",
					Type: "func",
					Func: &RFunc{
						IsFunc: ptr.To(true),
						InputParams: []*Type{
							{
								Name: "s",
								Type: "string",
							},
						},
						Returns: []*Type{
							{
								Type: "error",
							},
						},
					},
				},
			},
		},
	}
	tokens, err := Explore([]byte(`type Person struct {
	Name string
	Age int
	Job
}`))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	c := cur.NewCursor(tokens)

	strType := ExtractStructType(c)

	if strType == nil {
		t.Errorf("expected type not to be nil")
		return
	}
	if strType.Name != wants.Name {
		t.Errorf("unexpected name: wanted :%v ; got %v", *wants, *strType)
	}

}
