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
			wt.LogicBlocks = append(wt.LogicBlocks, ExtractFuncType(wt.Tokens))
		}
		wt.Tokens.Next()
	}
	return nil
}

func ExtractFuncType(c cur.Cursor[GoToken]) *Type {
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
		rcvTypes := ExtractParams(receiver)
		fn.Func.Receiver = rcvTypes[0]
		fn.Kind = TypeMethod
	}
	if generic != nil {
		fn.Generics = ExtractGenericType(generic)
	}
	if input != nil {
		fn.Func.InputParams = ExtractParams(input)
	}
	if returns != nil {
		fn.Func.Returns = ExtractParams(returns)
	}
	return fn
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
	if c.Cur().Tok == token.IDENT {
		returnStart := c.Pos()
		for c.Next().Tok != token.LBRACE {
			continue
		}
		c.Prev()
		var slice = []GoToken{
			{Tok: token.LPAREN},
		}
		slice = append(slice, c.Extract(returnStart, c.Pos())...)
		slice = append(slice, GoToken{Tok: token.RPAREN})
		returns = cur.NewCursor(slice)
	} else if c.Cur().Tok == token.LPAREN {
		returnStart := c.Pos()
		for c.Next().Tok != token.RPAREN {
			continue
		}
		returns = cur.NewCursor(c.Extract(returnStart, c.Pos()))
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
