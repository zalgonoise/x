package codegraph

import (
	"go/token"
	"strconv"

	cur "github.com/zalgonoise/cur"
	"github.com/zalgonoise/x/ptr"
)

// int
// *int
// context.Context
// *atomic.Value
// []int
// [3]int
// [length]int
// map[string]int
// map[*string][]int
// map[string]map[string]map[string]any
// []map[string]any

func ExtractParams(c cur.Cursor[GoToken]) []*Type {
	var (
		types = []*Type{}
	)

	t := &Type{}

	c.Tail()
	if c.PeekOffset(-1).Tok == token.LPAREN {
		return nil
	}
	for c.Pos() > 0 {
		c.Prev()
		switch c.Cur().Tok {
		case token.RBRACK:
			// generics: ...CreateFunc[T]
			if t.Type == "" {
				if t.Generics == nil {
					t.Generics = &RGeneric{}
				}
				t.Generics.Generics = ExtractGenericType(ExtractReverseCursor(c, TypeGenericParam))
				continue
			}
			// slices: []int
			if c.PeekOffset(-1).Tok == token.LBRACK {
				if t.Slice == nil {
					t.Slice = &RSlice{}
				}
				c.Prev()
				if c.PeekOffset(-1).Tok == token.MUL {
					t.Slice.IsPointer = ptr.To(true)
					c.Offset(-1)
				}
				continue
			}
			// arrays: [3]int
			if c.PeekOffset(-1).Tok == token.IDENT &&
				c.PeekOffset(-2).Tok == token.LBRACK &&
				c.PeekOffset(-3).Tok != token.MAP {
				if t.Slice == nil {
					t.Slice = &RSlice{}
				}
				v := c.Prev().Lit
				if n, err := strconv.Atoi(v); err != nil {
					t.Slice.Len = ptr.To(n)
				} else {
					t.Slice.LenName = ptr.To(v)
				}
				c.Prev()
				if c.PeekOffset(-1).Tok == token.MUL {
					t.Slice.IsPointer = ptr.To(true)
					c.Offset(-1)
				}
				continue
			}
			// maps: map[string]any
			if c.PeekOffset(-1).Tok == token.IDENT &&
				c.PeekOffset(-2).Tok == token.LBRACK &&
				c.PeekOffset(-3).Tok == token.MAP {
				if t.Map == nil {
					m := &RMap{
						Key:   c.Prev().Lit,
						Value: *t,
					}
					t.Map = m
				}
				c.Offset(-3)
				if c.PeekOffset(-1).Tok == token.MUL {
					t.Slice.IsPointer = ptr.To(true)
					c.Offset(-1)
				}
			}
		case token.ELLIPSIS:
			if t.Slice == nil {
				t.Slice = &RSlice{}
			}
			t.Slice.IsVariadic = ptr.To(true)
		case token.IDENT:
			if t.Type == "" {
				t.Type = c.Cur().Lit
			}
			switch c.PeekOffset(-1).Tok {
			case token.PERIOD:
				// package
				if c.PeekOffset(-2).Tok == token.IDENT {
					t.Package = ptr.To(c.Offset(-2).Lit)
				}
				if c.PeekOffset(-1).Tok == token.IDENT {
					t.Name = ptr.To(c.Prev().Lit)
				}
			case token.MUL:
				t.IsPointer = ptr.To(true)
				c.Prev()
			case token.RBRACK:
				continue
			}
			// all occurrences, for type, package, map-keyword and name
		case token.RPAREN:
			// functions: func(arg1, arg2, arg3) (ret1, ret2)
		case token.COMMA:
			// next argument
			types = append(types, ptr.Copy(t))
			t = &Type{}
		}
	}

	zero := &Type{}
	if t != zero {
		types = append(types, t)
	}

	return types
}
