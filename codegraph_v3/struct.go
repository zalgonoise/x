package codegraph

import (
	"go/token"

	cur "github.com/zalgonoise/cur"
	"github.com/zalgonoise/x/ptr"
)

func ExtractStruct(c cur.Cursor[GoToken]) *Type {
	name, fields := ExtractStructCursors(c)
	if len(fields) == 0 {
		return nil
	}

	str := &Type{
		Kind: TypeStruct,
		Name: name,
		Struct: &RStruct{
			IsStruct: ptr.To(true),
		},
	}

	for _, c := range fields {
		params := ExtractParamsReverse(c)
		if len(params) > 0 {
			str.Struct.Elems = append(str.Struct.Elems, params[0])
		}
	}
	return str
}

func ExtractStructCursors(c cur.Cursor[GoToken]) (
	name string, fields []cur.Cursor[GoToken],
) {
	if c.Cur().Tok != token.TYPE || c.PeekOffset(2).Tok != token.STRUCT {
		return "", nil
	}

	if c.Peek().Tok == token.IDENT {
		name = c.Next().Lit
	}
	if c.Offset(2).Tok == token.LBRACE {
		for c.Cur().Tok != token.RBRACE {
			fieldStart := c.Pos()
			for c.Cur().Tok != token.SEMICOLON {
				c.Next()
			}
			fields = append(fields, cur.NewCursor(c.Extract(fieldStart, c.Pos())))
			c.Next()
		}
	}
	return

}

func ExtractObjectType(c cur.Cursor[GoToken]) *Type {
	kind, fields := ExtractObjectTypeCursors(c)
	if kind == 0 {
		return nil
	}

	switch kind {
	case TypeStruct:
		str := &Type{
			Type: "struct",
			Kind: TypeStruct,
			Struct: &RStruct{
				IsStruct: ptr.To(true),
			},
		}

		for _, f := range fields {
			params := ExtractParamsReverse(f)
			if len(params) > 0 {
				str.Struct.Elems = append(str.Struct.Elems, params[0])
			}
		}
		return str
	case TypeInterface:

	}
	return nil
}

func ExtractObjectTypeCursors(c cur.Cursor[GoToken]) (
	kind LogicBlockKind, fields []cur.Cursor[GoToken],
) {
	rollback := c.Pos()
	for c.Cur().Tok != token.LBRACE &&
		c.Pos() >= 0 {
		c.Prev()
	}
	if c.PeekOffset(-1).Tok == token.STRUCT {
		kind = TypeStruct
	} else if c.PeekOffset(-1).Tok == token.INTERFACE {
		kind = TypeInterface
	} else {
		c.Idx(rollback)
		return 0, nil
	}

	s := c.Pos()
	defer c.Idx(s)
	if c.Cur().Tok == token.LBRACE {
		c.Next()
		for c.Cur().Tok != token.RBRACE {
			fieldStart := c.Pos()
			for c.Cur().Tok != token.SEMICOLON {
				c.Next()
			}
			fields = append(fields, cur.NewCursor(WrapType(c.Extract(fieldStart, c.Pos()), TypeFuncParam)))
			c.Next()
		}
	}
	return

}
