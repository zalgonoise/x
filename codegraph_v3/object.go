package codegraph

import (
	"fmt"
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

	for _, f := range fields {
		params := ExtractParamsReverse(f)
		if len(params) > 0 {
			str.Struct.Elems = append(str.Struct.Elems, params[0])
		}
	}
	return str
}

func ExtractStructCursors(c cur.Cursor[GoToken]) (
	name string, fields []cur.Cursor[GoToken],
) {
	if c.Cur().Tok != token.TYPE && c.PeekOffset(2).Tok != token.STRUCT {
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
			fields = append(fields, cur.NewCursor(WrapType(c.Extract(fieldStart, c.Pos()), TypeFuncParam)))
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
		itf := &Type{
			Type: "interface",
			Kind: TypeInterface,
			Interface: &RInterface{
				IsInterface: ptr.To(true),
			},
		}

		for _, f := range fields {
			method := ExtractMethod(f)
			if method != nil {
				itf.Interface.Methods = append(itf.Interface.Methods, method)
			}
		}
		return itf

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

func ExtractInterface(c cur.Cursor[GoToken]) *Type {
	name, methods := ExtractInterfaceCursors(c)

	itf := &Type{
		Kind: TypeInterface,
		Type: "interface",
		Name: name,
		Interface: &RInterface{
			IsInterface: ptr.To(true),
		},
	}

	for _, f := range methods {
		// TODO: needs its own extractor
		method := ExtractMethod(f)
		if methods != nil {
			itf.Interface.Methods = append(itf.Interface.Methods, method)
		}
	}
	return itf
}

func ExtractInterfaceCursors(c cur.Cursor[GoToken]) (
	name string, methods []cur.Cursor[GoToken],
) {
	if c.Cur().Tok != token.TYPE && c.PeekOffset(2).Tok != token.INTERFACE {
		return "", nil
	}
	if c.Peek().Tok == token.IDENT {
		name = c.Next().Lit
		c.Offset(2)
	}
	if c.Cur().Tok == token.LBRACE {
		c.Next()
		for c.Cur().Tok != token.RBRACE {
			methodStart := c.Pos()
			for c.Cur().Tok != token.SEMICOLON {
				c.Next()
			}
			methods = append(methods, cur.NewCursor(c.Extract(methodStart, c.Pos())))
			c.Next()
		}
	}
	return
}

func ExtractMethod(c cur.Cursor[GoToken]) *Type {
	name, input, returns := ExtractMethodCursors(c)

	fmt.Println(name, "\n", input, "\n", returns)

	var (
		inputParams  = []*Type{}
		returnParams = []*Type{}
	)
	if input != nil {
		inputParams = ExtractParamsReverse(input)
	}
	if returns != nil {
		returnParams = ExtractParamsReverse(returns)
	}
	method := &Type{
		Name: name,
		Type: "method",
		Kind: TypeMethod,
		Func: &RFunc{
			IsFunc:      ptr.To(true),
			InputParams: inputParams,
			Returns:     returnParams,
		},
	}
	return method
}

func ExtractMethodCursors(c cur.Cursor[GoToken]) (
	name string, input, returns cur.Cursor[GoToken],
) {
	c.Head()
	if c.Cur().Tok == token.LPAREN {
		c.Next()
	}
	if c.Cur().Tok == token.IDENT {
		name = c.Cur().Lit
		c.Next()
	}
	if c.Cur().Tok == token.LPAREN {
		inputStart := c.Pos()
		for c.Cur().Tok != token.RPAREN {
			c.Next()
		}
		input = cur.NewCursor(c.Extract(inputStart, c.Pos()))
		c.Next()
	}
	if c.Cur().Tok == token.LPAREN {
		returnStart := c.Pos()
		for c.Cur().Tok != token.RPAREN {
			c.Next()
		}
		returns = cur.NewCursor(c.Extract(returnStart, c.Pos()))
	} else {
		returnStart := c.Pos()
		for c.Cur().Tok != token.SEMICOLON {
			c.Next()
		}
		returns = cur.NewCursor(WrapType(c.Extract(returnStart, c.Pos()), TypeFuncParam))
	}
	return
}
