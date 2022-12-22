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
	// rewind cursor until it hits the end of the parenthesis
	for c.Prev().Tok != token.RPAREN && c.Pos() != 0 {
		continue
	}
	if c.PeekOffset(-1).Tok == token.LPAREN || c.Pos() == 0 {
		return nil
	}
	for c.Pos() > 0 {
		c.Prev()
		switch c.Cur().Tok {
		case token.RBRACK:
			// generics: ...CreateFunc[T]
			if t.Type == "" {
				if t.Generics == nil {
					t.Generics = []*Type{}
				}
				t.Generics = ExtractGenericType(ExtractReverseCursor(c, TypeGenericParam))
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
			if (c.PeekOffset(-1).Tok == token.IDENT || c.PeekOffset(-1).Tok == token.INT) &&
				c.PeekOffset(-2).Tok == token.LBRACK &&
				c.PeekOffset(-3).Tok != token.MAP {
				if t.Slice == nil {
					t.Slice = &RSlice{}
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
		case token.ELLIPSIS:
			if t.Slice == nil {
				t.Slice = &RSlice{}
			}
			t.Slice.IsVariadic = ptr.To(true)
		case token.IDENT:
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
			case token.RBRACK:
				continue
			}
			// all occurrences, for type, package, map-keyword and name
		case token.RPAREN:
			// functions: func(arg1, arg2, arg3) (ret1, ret2)
		case token.COMMA:
			types = append(types, ptr.Copy(t))
			if c.PeekOffset(-1).Tok == token.IDENT {
				if c.PeekOffset(-2).Tok == token.COMMA || c.PeekOffset(-2).Tok == token.LPAREN {
					t.Name = c.Prev().Lit
				} else {
					*t = Type{}
				}
			}

		case token.LPAREN:
			zero := &Type{}
			if t != zero {
				types = append(types, t)
			}
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
