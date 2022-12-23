package codegraph

import (
	"fmt"
	"go/token"
	"strconv"

	cur "github.com/zalgonoise/cur"
	"github.com/zalgonoise/x/ptr"
)

// int
// *int
// n int
// context.Context
// *atomic.Value
// []int
// [3]int
// [length]int
// map[string]int
// map[*string][]int
// map[string]map[string]map[string]any
// []map[string]any
// func(s string) error
// func(x, y int) (int, error)
// func(x) func(x) bool

// func ExtractParams(c cur.Cursor[GoToken]) []*Type {
// 	c.Head()
// 	// roll cursor until it hits the start of the parenthesis
// 	for c.Cur().Tok != token.LPAREN && c.Pos() != c.Len()-1 {
// 		c.Next()
// 	}
// 	if c.Peek().Tok == token.RPAREN || c.Pos() == c.Len()-1 {
// 		return nil
// 	}

// 	var (
// 		types = []*Type{}
// 		t     = &Type{}
// 	)

// 	for c.Pos() < c.Len() {
// 		c.Next()
// 		switch c.Cur().Tok {
// 		case token.IDENT:
// 			tokenIDENT(c, t)
// 		case token.MUL:
// 			tokenMUL(c, t)
// 		case token.LBRACK:
// 			tokenLBRACK(c, t)
// 		case token.MAP:
// 		case token.FUNC:
// 		case token.STRUCT:
// 		case token.INTERFACE:
// 		case token.COMMA:
// 		}
// 	}
// }

// func tokenIDENT(c cur.Cursor[GoToken], t *Type) {
// 	// check if it's a name or a type
// 	if c.Peek().Tok == token.IDENT {
// 		t.Name = c.Cur().Lit
// 		t.Type = c.Next().Lit
// 	} else {
// 		t.Type = c.Cur().Lit
// 	}
// 	if c.Peek().Tok == token.PERIOD &&
// 	c.PeekOffset(2).Tok == token.IDENT {
// 		t.Package = t.Type
// 		t.Type = c.Offset(2).Lit
// 	}
// }

// func tokenMUL(c cur.Cursor[GoToken], t *Type) {
// 	t.IsPointer = ptr.To(true)
// }

// func tokenLBRACK(c cur.Cursor[GoToken], t *Type) {
// 	switch c.Peek().Tok {
// 	case token.RBRACK:
// 		if t.Type != "" {}
// 	}
// }

func ExtractParamsReverse(c cur.Cursor[GoToken]) []*Type {
	c.Tail()
	// rewind cursor until it hits the end of the parenthesis
	for c.Cur().Tok != token.RPAREN && c.Pos() != 0 {
		c.Prev()
	}
	if c.PeekOffset(-1).Tok == token.LPAREN || c.Pos() == 0 {
		return nil
	}

	var (
		types = []*Type{}
		t     = &Type{}
	)

	for c.Pos() > 0 {
		c.Prev()
		switch c.Cur().Tok {
		case token.RBRACK:
			// generics: ...CreateFunc[T]
			// slices / arrays: []int, [3]uint
			// maps: map[string]any
			handleRBRACK(c, t)
		case token.ELLIPSIS:
			handleELLIPSIS(c, t)
		case token.IDENT:
			// all occurrences, for type, package, map-keyword and name
			handleIDENT(c, t)
		case token.RPAREN:
			// functions: func(arg1, arg2, arg3) (ret1, ret2)
			handleFUNC(c, t)
		case token.RBRACE:
			// structs, interfaces: struct{} ; interface{ String() string }
			// handleSTRUCTorINTERFACE
		case token.COMMA:
			handleCOMMA(c, t, &types)
		case token.LPAREN:
			handleLPAREN(c, t, &types)
		}
	}

	if len(types) > 1 {
		output := make([]*Type, len(types), len(types))
		for idx, t := range types {
			output[len(types)-1-idx] = t
		}
		return output
	}
	return types
}

func handleRBRACK(c cur.Cursor[GoToken], t *Type) {
	if t.Type == "" {
		if t.Generics == nil {
			t.Generics = []*Type{}
		}
		t.Generics = ExtractGenericType(ExtractReverseCursor(c, TypeGenericParam))
		return
	}
	// slices: []int
	if c.PeekOffset(-1).Tok == token.LBRACK {
		if t.Slice == nil {
			t.Slice = &RSlice{
				IsSlice: ptr.To(true),
			}
		}
		c.Prev()
		if c.PeekOffset(-1).Tok == token.MUL {
			t.Slice.IsPointer = ptr.To(true)
			c.Offset(-1)
		}
		return
	}
	// arrays: [3]int
	if (c.PeekOffset(-1).Tok == token.IDENT || c.PeekOffset(-1).Tok == token.INT) &&
		c.PeekOffset(-2).Tok == token.LBRACK &&
		c.PeekOffset(-3).Tok != token.MAP {
		if t.Slice == nil {
			t.Slice = &RSlice{
				IsSlice: ptr.To(true),
			}
		}
		if c.Prev().Tok == token.INT {
			n, err := strconv.Atoi(c.Cur().Lit)
			if err == nil {
				t.Slice.Len = ptr.To(n)
			}
		} else {
			t.Slice.LenName = ptr.To(c.Cur().Lit)
		}
		c.Prev()
		if c.PeekOffset(-1).Tok == token.MUL {
			t.Slice.IsPointer = ptr.To(true)
			c.Offset(-1)
		}
		return
	}
	// maps: map[string]any
	if c.PeekOffset(-1).Tok == token.IDENT &&
		c.PeekOffset(-2).Tok == token.LBRACK &&
		c.PeekOffset(-3).Tok == token.MAP {
		if t.Map == nil {
			m := &RMap{
				IsMap: ptr.To(true),
				Key:   c.Prev().Lit,
				Value: *t,
			}
			t.Map = m
			t.Type = "map"
		}
		c.Offset(-2)
		if c.PeekOffset(-1).Tok == token.MUL {
			t.Map.IsPointer = ptr.To(true)
			c.Offset(-1)
		}
		if c.PeekOffset(-1).Tok == token.IDENT {
			t.Name = c.Prev().Lit
		}
	}
}

func handleELLIPSIS(c cur.Cursor[GoToken], t *Type) {
	if t.Slice == nil {
		t.Slice = &RSlice{
			IsSlice: ptr.To(true),
		}
	}
	t.Slice.IsVariadic = ptr.To(true)
}

func handleIDENT(c cur.Cursor[GoToken], t *Type) {
	if t.Type == "" {
		t.Type = c.Cur().Lit
	} else if t.Name == "" {
		t.Name = c.Cur().Lit // adds name for slices and maps
	}
	switch c.PeekOffset(-1).Tok {
	case token.IDENT:
		// named type
		t.Name = c.Offset(-1).Lit
	case token.PERIOD:
		// package
		if c.PeekOffset(-2).Tok == token.IDENT {
			t.Package = ptr.To(c.Offset(-2).Lit)
		}
		if c.PeekOffset(-1).Tok == token.IDENT {
			t.Name = c.Prev().Lit
		}
	case token.MUL:
		t.IsPointer = ptr.To(true)
		c.Prev()
	case token.RBRACK, token.COMMA:
		return
	}
}

func handleCOMMA(c cur.Cursor[GoToken], t *Type, types *[]*Type) {
	*types = append(*types, ptr.Copy(t))
	if c.PeekOffset(-1).Tok == token.IDENT {
		if t.Name != "" &&
			(c.PeekOffset(-2).Tok == token.COMMA || c.PeekOffset(-2).Tok == token.LPAREN) {
			t.Name = c.Prev().Lit
		} else {
			*t = Type{}
		}
	}
}

func handleLPAREN(c cur.Cursor[GoToken], t *Type, types *[]*Type) {
	if t.Type != "" {
		*types = append(*types, t)
	}
}

func handleFUNC(c cur.Cursor[GoToken], t *Type) {
	fn := ExtractFuncType(c)
	if t.Type != "" || t.Name != "" || t.Func != nil || t.Map != nil || t.Slice != nil {
		fn.Func.Returns = []*Type{ptr.Copy(t)}
	}
	fn.Type = "func"
	if c.PeekOffset(-1).Tok == token.IDENT {
		fn.Name = c.Prev().Lit
	}
	fmt.Println(fn)
	*t = *fn
}
