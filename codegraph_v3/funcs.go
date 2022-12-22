package codegraph

import (
	"go/token"

	cur "github.com/zalgonoise/cur"
	"github.com/zalgonoise/x/ptr"
)

func (wt *WithTokens) Func() error {
	if wt.LogicBlocks == nil {
		wt.LogicBlocks = []*Type{}
	}

	wt.Tokens.Head()

	for i := wt.Tokens.Pos(); i < wt.Tokens.Len(); i++ {
		if wt.Tokens.Cur().Tok == token.FUNC {
			wt.LogicBlocks = append(wt.LogicBlocks, ExtractFunc(wt.Tokens))
		}
		wt.Tokens.Next()
	}
	return nil
}

func ExtractFunc(c cur.Cursor[GoToken]) *Type {
	name, receiver, generic, input, returns, logic := ExtractFuncCursors(c)
	if name == "" || logic == nil {
		return nil
	}

	fn := &Type{
		Kind: TypeFunction,
		Name: name,
		Func: &RFunc{
			IsFunc: ptr.To(true),
		},
	}
	if receiver != nil {
		rcvTypes := ExtractParamsReverse(receiver)
		fn.Func.Receiver = rcvTypes[0]
		fn.Kind = TypeMethod
	}
	if generic != nil {
		fn.Generics = ExtractGenericType(generic)
	}
	if input != nil {
		fn.Func.InputParams = ExtractParamsReverse(input)
	}
	if returns != nil {
		fn.Func.Returns = ExtractParamsReverse(returns)
	}
	return fn
}

func ExtractFuncType(c cur.Cursor[GoToken]) *Type {
	var returnParams []*Type = nil
	name, input, returns := ExtractFuncTypeCursors(c)
	if returns != nil {
		returnParams = ExtractParamsReverse(returns)
	}

	fnType := &Type{
		Kind: TypeFunction,
		Name: name,
		Func: &RFunc{
			IsFunc:      ptr.To(true),
			InputParams: ExtractParamsReverse(input),
			Returns:     returnParams,
		},
	}
	return fnType
}

func ExtractFuncCursors(c cur.Cursor[GoToken]) (
	name string,
	receiver, generic, input, returns, logic cur.Cursor[GoToken],
) {
	if c.Cur().Tok != token.FUNC {
		return "", nil, nil, nil, nil, nil
	}

	var (
		fnLogicStart = -1
		braceLvl     = 0
	)

	// handle receivers
	if c.Peek().Tok == token.LPAREN {
		c.Next()
		receiverStart := c.Pos()
		for c.Next().Tok != token.RPAREN {
			continue
		}
		receiver = cur.NewCursor(c.Extract(receiverStart, c.Pos()))
	}
	if c.Peek().Tok == token.IDENT {
		name = c.Next().Lit
	}
	if c.Peek().Tok == token.LBRACK {
		c.Next()
		genericStart := c.Pos()
		for c.Next().Tok != token.RBRACK {
			continue
		}
		generic = cur.NewCursor(c.Extract(genericStart, c.Pos()))
	}
	if c.Peek().Tok == token.LPAREN {
		c.Next()
		inputParamStart := c.Pos()
		for c.Next().Tok != token.RPAREN {
			continue
		}
		input = cur.NewCursor(c.Extract(inputParamStart, c.Pos()))
	}
	c.Next()

	if c.Cur().Tok == token.LPAREN {
		returnStart := c.Pos()
		for c.Next().Tok != token.RPAREN {
			continue
		}
		returns = cur.NewCursor(c.Extract(returnStart, c.Pos()))
	} else {
		switch c.Cur().Tok {
		case token.IDENT, token.MUL, token.MAP, token.FUNC, token.STRUCT, token.INTERFACE, token.LBRACK:
			returnStart := c.Pos()
			for t := c.Peek().Tok; t != token.LBRACE && t != token.SEMICOLON && t != token.EOF; {
				c.Next()
			}
			returns = cur.NewCursor(WrapType(c.Extract(returnStart, c.Pos()), TypeFuncParam))
		}
	}
	if c.Next().Tok == token.LBRACE {
		fnLogicStart = c.Pos()
		braceLvl++
	}
	for c.Next().Tok != token.RBRACE && braceLvl == 1 {
		if c.Cur().Tok == token.LBRACE {
			braceLvl++
		}
	}
	logic = cur.NewCursor(c.Extract(fnLogicStart, c.Pos()))

	return
}

func ExtractFuncTypeCursors(c cur.Cursor[GoToken]) (
	name string,
	input, returns cur.Cursor[GoToken],
) {
	// rewind to func start
	rollback := c.Pos()
	for c.Cur().Tok != token.FUNC && c.Pos() >= 0 {
		c.Prev()
	}
	if c.Cur().Tok != token.FUNC ||
		(c.Cur().Tok == token.IDENT && c.Peek().Tok == token.FUNC) {
		c.Idx(rollback)
		return "", nil, nil
	}

	s := c.Pos()
	defer c.Idx(s)
	if c.Cur().Tok == token.IDENT {
		name = c.Cur().Lit
		c.Next()
	}
	if c.Cur().Tok == token.FUNC && c.Peek().Tok == token.LPAREN {
		c.Next()
		inputStart := c.Pos()
		for c.Cur().Tok != token.RPAREN {
			c.Next()
		}
		input = cur.NewCursor(c.Extract(inputStart, c.Pos()))
		c.Next()
	}

	// switch c.Cur().Tok {
	// case token.LPAREN:
	if c.Cur().Tok == token.LPAREN {
		returnStart := c.Pos()
		for c.Cur().Tok != token.RPAREN {
			c.Next()
		}
		returns = cur.NewCursor(c.Extract(returnStart, c.Pos()))
	}
	return

	// case token.IDENT, token.MUL, token.MAP, token.FUNC, token.STRUCT, token.INTERFACE, token.LBRACK:
	// 	returnStart := c.Pos()
	// 	for t := c.Peek().Tok; t != token.LBRACE && t != token.SEMICOLON && t != token.EOF; {
	// 		fmt.Println(c.Cur().Tok.String(), c.Cur().Lit, c.Peek().Tok.String())

	// 		c.Next()
	// 	}
	// 	c.Prev()
	// 	returns = cur.NewCursor(WrapType(c.Extract(returnStart, c.Pos()), TypeFuncParam))
	// }

	// return
}

func ExtractGenericType(c cur.Cursor[GoToken]) []*Type {
	generics := []*Type{}
	gen := &Type{}

	c.Tail()
	for c.Pos() > 0 {
		c.Prev()
		switch c.Cur().Tok {
		case token.IDENT:
			if gen.Type == "" {
				gen.Type = c.Cur().Lit
			} else {
				gen.Name = c.Cur().Lit
			}

		case token.COMMA:
			generics = append(generics, ptr.Copy(gen))
			if c.PeekOffset(-1).Tok == token.IDENT {
				if c.PeekOffset(-2).Tok == token.COMMA || c.PeekOffset(-2).Tok == token.LBRACK {
					gen.Name = c.Prev().Lit
				} else {
					*gen = Type{}
				}
			}
		case token.LBRACK:
			zero := &Type{}
			if gen != zero {
				generics = append(generics, gen)
			}
		}
	}
	if len(generics) > 1 {
		output := make([]*Type, len(generics), len(generics))
		for idx, t := range generics {
			output[len(generics)-1-idx] = t
		}
		return output
	}
	return generics
}
