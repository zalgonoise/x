package codegraph

import (
	"fmt"
	"go/token"

	cur "github.com/zalgonoise/cur"
	"github.com/zalgonoise/x/ptr"
)

func (wt *WithTokens) Func() error {
	if wt.LogicBlocks == nil {
		wt.LogicBlocks = []*LogicBlock{}
	}

	for i := wt.Tokens.Pos(); i < wt.Tokens.Len(); i++ {
		if wt.Tokens.Cur().Tok == token.FUNC {
			fn := ExtractCursor(wt.Tokens, TypeFunction)
			lb := extractFunc(fn)
			wt.LogicBlocks = append(wt.LogicBlocks, lb)
		}
		wt.Tokens.Next()
	}
	return nil
}

func extractFunc(c cur.Cursor[GoToken]) *LogicBlock {
	lb := &LogicBlock{
		Kind: TypeFunction,
	}

	if c.Cur().Tok == token.FUNC &&
		c.Peek().Tok == token.IDENT {
		lb.Name = ptr.To(c.Next().Lit)
	} else if c.Cur().Tok == token.FUNC &&
		c.Peek().Tok == token.LPAREN {
		c.Next()
		recv := extractReceiver(ExtractCursor(c, TypeFuncParam))
		lb.Receiver = recv
		if c.Peek().Tok == token.IDENT {
			lb.Name = ptr.To(c.Next().Lit)
		}
	}

	c.Next()
	if c.Cur().Tok == token.LBRACK {
		gen := extractGenerics(ExtractCursor(c, TypeGenericParam))
		lb.Generics = gen
		c.Next()
	}
	if c.Cur().Tok == token.LPAREN {
		inputParam := extractParams(ExtractCursor(c, TypeFuncParam))
		for _, p := range inputParam {
			fmt.Println(*p.Type)
		}

		lb.InputParams = inputParam
		c.Next()
	}

	// for c.Cur().Tok != token.LBRACE {
	// 	switch c.Cur().Tok {
	// 	case token.LPAREN:
	// 		// output param group
	// 	case token.IDENT:
	// 		// output param
	// 	}
	// }

	return lb
}

func extractReceiver(c cur.Cursor[GoToken]) *Identifier {
	var id = &Identifier{}

	if c.PeekOffset(1).Tok == token.IDENT &&
		c.PeekOffset(2).Tok == token.IDENT {
		id.Name = ptr.To(c.Next().Lit)
		id.Type = c.Next().Lit
	} else if c.PeekOffset(1).Tok == token.IDENT &&
		c.PeekOffset(2).Tok == token.MUL &&
		c.PeekOffset(3).Tok == token.IDENT {
		id.Name = ptr.To(c.Next().Lit)
		c.Next()
		id.IsPointer = ptr.To(true)
		id.Type = c.Next().Lit
	} else if c.PeekOffset(1).Tok == token.MUL &&
		c.PeekOffset(2).Tok == token.IDENT {
		c.Next()
		id.IsPointer = ptr.To(true)
		id.Type = c.Next().Lit
	} else {
		id.Type = c.Next().Lit
	}

	if c.Peek().Tok == token.LBRACK {
		c.Next()
		gen := extractGenerics(ExtractCursor(c, TypeGenericParam))
		id.GenericTypes = gen
	}
	return id
}

func extractGenerics(c cur.Cursor[GoToken]) []*Identifier {
	var ids = []*Identifier{}
	var nameIDs = []string{}

	for i := c.Pos(); i < c.Len(); i++ {
		c.Next()
		if c.Cur().Tok == token.IDENT &&
			c.Peek().Tok == token.COMMA {
			nameIDs = append(nameIDs, c.Cur().Lit)
			c.Next()
			continue
		}
		if c.Cur().Tok == token.IDENT &&
			c.Peek().Tok == token.IDENT {
			for _, name := range nameIDs {
				ids = append(ids, &Identifier{
					Name: &name,
					Type: c.Peek().Lit,
				})
			}
			nameIDs = nameIDs[:0]
			c.Next()
			continue
		}
		if c.Cur().Tok == token.COMMA {
			c.Next()
		}
	}

	return ids
}

func extractParams(c cur.Cursor[GoToken]) []*LogicBlock {
	var lb = []*LogicBlock{}

	for i := c.Pos(); i < c.Len(); i++ {
		c.Next()

		if c.Cur().Tok == token.IDENT &&
			c.PeekOffset(1).Tok == token.PERIOD &&
			c.PeekOffset(2).Tok == token.IDENT {

			lb = append(lb, &LogicBlock{
				Package: c.Cur().Lit,
				Type:    ptr.To(c.Offset(2).Lit),
			})
			i += 2
			continue
		}
		if c.Cur().Tok == token.IDENT &&
			c.PeekOffset(1).Tok == token.IDENT &&
			c.PeekOffset(2).Tok == token.PERIOD &&
			c.PeekOffset(3).Tok == token.IDENT {
			lb = append(lb, &LogicBlock{
				Name:    ptr.To(c.Cur().Lit),
				Package: c.Peek().Lit,
				Type:    ptr.To(c.Offset(3).Lit),
			})
			i += 3
			continue
		}
		if c.Cur().Tok == token.IDENT &&
			c.PeekOffset(1).Tok == token.IDENT {
			lb = append(lb, &LogicBlock{
				Name: ptr.To(c.Cur().Lit),
				Type: ptr.To(c.Offset(1).Lit),
			})
			i += 1
			continue
		}
		if c.Cur().Tok == token.IDENT &&
			c.PeekOffset(1).Tok == token.MUL &&
			c.PeekOffset(2).Tok == token.IDENT &&
			c.PeekOffset(3).Tok == token.PERIOD &&
			c.PeekOffset(4).Tok == token.IDENT {
			lb = append(lb, &LogicBlock{
				Name:      ptr.To(c.Cur().Lit),
				Package:   c.PeekOffset(2).Lit,
				Type:      ptr.To(c.Offset(4).Lit),
				IsPointer: ptr.To(true),
			})
			i += 4
			continue
		}
		if c.Cur().Tok == token.IDENT &&
			c.PeekOffset(1).Tok == token.MUL &&
			c.PeekOffset(2).Tok == token.IDENT {
			lb = append(lb, &LogicBlock{
				Name:      ptr.To(c.Cur().Lit),
				Type:      ptr.To(c.Offset(2).Lit),
				IsPointer: ptr.To(true),
			})
			i += 2
			continue
		}
		if c.Cur().Tok == token.MUL &&
			c.PeekOffset(1).Tok == token.IDENT {
			lb = append(lb, &LogicBlock{
				Type:      ptr.To(c.Offset(1).Lit),
				IsPointer: ptr.To(true),
			})
			i += 1
			continue
		}
		if c.Cur().Tok == token.COMMA {
			continue
		}
		if c.Cur().Tok == token.FUNC {
			fn := extractFunc(ExtractCursor(c, TypeFunction))
			lb = append(lb, fn)
			continue
		}
		if c.Cur().Tok == token.IDENT && c.Cur().Lit != "" {
			lb = append(lb, &LogicBlock{
				Type: ptr.To(c.Cur().Lit),
			})
		}
	}

	return lb
}
