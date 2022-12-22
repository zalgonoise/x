package codegraph

import (
	"go/token"

	cur "github.com/zalgonoise/cur"
	"github.com/zalgonoise/x/ptr"
)

func ExtractStructType(c cur.Cursor[GoToken]) *Type {
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
