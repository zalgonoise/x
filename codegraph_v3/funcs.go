package codegraph

import (
	"go/token"

	cur "github.com/zalgonoise/cur"
	"github.com/zalgonoise/x/ptr"
)

func (wt *WithTokens) Func() error {
	// if wt.LogicBlocks == nil {
	// 	wt.LogicBlocks = []*LogicBlock{}
	// }

	// for i := wt.Tokens.Pos(); i < wt.Tokens.Len(); i++ {
	// 	if wt.Tokens.Cur().Tok == token.FUNC {
	// 		fn := ExtractCursor(wt.Tokens, TypeFunction)
	// 		lb := extractFunc(fn)
	// 		wt.LogicBlocks = append(wt.LogicBlocks, lb)
	// 	}
	// 	wt.Tokens.Next()
	// }
	return nil
}

// func extractFunc(c cur.Cursor[GoToken]) *LogicBlock {
// 	lb := &LogicBlock{
// 		Kind: TypeFunction,
// 	}

// 	if c.Cur().Tok == token.FUNC &&
// 		c.Peek().Tok == token.IDENT {
// 		lb.Name = ptr.To(c.Next().Lit)
// 	} else if c.Cur().Tok == token.FUNC &&
// 		c.Peek().Tok == token.LPAREN {
// 		c.Next()
// 		recv := extractReceiver(ExtractCursor(c, TypeFuncParam))
// 		lb.Receiver = recv
// 		lb.Kind = TypeMethod
// 		if c.Peek().Tok == token.IDENT {
// 			lb.Name = ptr.To(c.Next().Lit)
// 		}
// 	}

// 	c.Next()
// 	if c.Cur().Tok == token.LBRACK {
// 		gen := extractGenerics(ExtractCursor(c, TypeGenericParam))
// 		lb.Generics = gen
// 		c.Next()
// 	}
// 	if c.Cur().Tok == token.LPAREN {
// 		inputParam := extractParams(ExtractCursor(c, TypeFuncParam), true)
// 		lb.InputParams = inputParam
// 	}

// 	for c.Next().Tok != token.LBRACE {
// 		switch c.Cur().Tok {
// 		case token.MUL:
// 			if c.Peek().Tok == token.IDENT &&
// 				c.PeekOffset(2).Tok == token.PERIOD &&
// 				c.PeekOffset(3).Tok == token.IDENT {
// 				if lb.ReturnParams == nil {
// 					lb.ReturnParams = []*LogicBlock{}
// 				}
// 				lb.ReturnParams = append(lb.ReturnParams, &LogicBlock{
// 					IsPointer: ptr.To(true),
// 					Package:   c.Peek().Lit,
// 					Type:      ptr.To(c.Offset(3).Lit),
// 					Kind:      TypeFuncReturn,
// 				})
// 				continue
// 			}
// 			if c.Peek().Tok == token.IDENT {
// 				if lb.ReturnParams == nil {
// 					lb.ReturnParams = []*LogicBlock{}
// 				}
// 				lb.ReturnParams = append(lb.ReturnParams, &LogicBlock{
// 					IsPointer: ptr.To(true),
// 					Type:      ptr.To(c.Next().Lit),
// 					Kind:      TypeFuncReturn,
// 				})
// 				continue
// 			}

// 		case token.IDENT:
// 			if c.Peek().Tok == token.PERIOD &&
// 				c.PeekOffset(2).Tok == token.IDENT {
// 				if lb.ReturnParams == nil {
// 					lb.ReturnParams = []*LogicBlock{}
// 				}
// 				lb.ReturnParams = append(lb.ReturnParams, &LogicBlock{
// 					Package: c.Cur().Lit,
// 					Type:    ptr.To(c.Offset(2).Lit),
// 					Kind:    TypeFuncReturn,
// 				})
// 				continue
// 			}
// 			if c.Cur().Lit != "" {
// 				if lb.ReturnParams == nil {
// 					lb.ReturnParams = []*LogicBlock{}
// 				}
// 				lb.ReturnParams = append(lb.ReturnParams, &LogicBlock{
// 					Type: ptr.To(c.Cur().Lit),
// 					Kind: TypeFuncReturn,
// 				})
// 				continue
// 			}

// 		case token.LPAREN:
// 			returnParams := extractParams(ExtractCursor(c, TypeFuncParam), false)
// 			lb.ReturnParams = returnParams
// 		}
// 	}

// 	return lb
// }

// func extractReceiver(c cur.Cursor[GoToken]) *Identifier {
// 	var id = &Identifier{
// 		Kind: TypeReceiver,
// 	}

// 	if c.PeekOffset(1).Tok == token.IDENT &&
// 		c.PeekOffset(2).Tok == token.IDENT {
// 		id.Name = ptr.To(c.Next().Lit)
// 		id.Type = c.Next().Lit
// 	} else if c.PeekOffset(1).Tok == token.IDENT &&
// 		c.PeekOffset(2).Tok == token.MUL &&
// 		c.PeekOffset(3).Tok == token.IDENT {
// 		id.Name = ptr.To(c.Next().Lit)
// 		c.Next()
// 		id.IsPointer = ptr.To(true)
// 		id.Type = c.Next().Lit
// 	} else if c.PeekOffset(1).Tok == token.MUL &&
// 		c.PeekOffset(2).Tok == token.IDENT {
// 		c.Next()
// 		id.IsPointer = ptr.To(true)
// 		id.Type = c.Next().Lit
// 	} else {
// 		id.Type = c.Next().Lit
// 	}

// 	if c.Peek().Tok == token.LBRACK {
// 		c.Next()
// 		gen := extractGenerics(ExtractCursor(c, TypeGenericParam))
// 		id.GenericTypes = gen
// 	}
// 	return id
// }

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

	return generics
}

// func extractGenerics(c cur.Cursor[GoToken]) []*Identifier {
// 	var ids = []*Identifier{}
// 	var nameIDs = []string{}

// 	for i := c.Pos(); i < c.Len(); i++ {
// 		c.Next()
// 		if c.Cur().Tok == token.IDENT &&
// 			c.Peek().Tok == token.COMMA {
// 			nameIDs = append(nameIDs, c.Cur().Lit)
// 			c.Next()
// 			continue
// 		}
// 		if c.Cur().Tok == token.IDENT &&
// 			c.Peek().Tok == token.IDENT {
// 			for _, name := range nameIDs {
// 				ids = append(ids, &Identifier{
// 					Name: &name,
// 					Type: c.Peek().Lit,
// 					Kind: TypeGenericParam,
// 				})
// 			}
// 			nameIDs = nameIDs[:0]
// 			ids = append(ids, &Identifier{
// 				Name: ptr.To(c.Cur().Lit),
// 				Type: c.Peek().Lit,
// 				Kind: TypeGenericParam,
// 			})
// 			c.Next()
// 			continue
// 		}
// 		if c.Cur().Tok == token.IDENT &&
// 			c.Peek().Tok == token.RBRACK {
// 			ids = append(ids, &Identifier{
// 				Type: c.Cur().Lit,
// 				Kind: TypeGenericParam,
// 			})
// 			c.Next()
// 			continue
// 		}
// 		if c.Cur().Tok == token.COMMA {
// 			c.Next()
// 		}
// 	}

// 	return ids
// }

// func extractParams(c cur.Cursor[GoToken], isInput bool) []*LogicBlock {
// 	var (
// 		lb                    = []*LogicBlock{}
// 		paramT LogicBlockKind = TypeFuncReturn
// 	)
// 	if isInput {
// 		paramT = TypeFuncParam
// 	}

// 	for i := c.Pos(); i < c.Len(); i++ {
// 		c.Next()

// 		// "context.Context"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.PERIOD &&
// 			c.PeekOffset(2).Tok == token.IDENT {

// 			lb = append(lb, &LogicBlock{
// 				Package: c.Cur().Lit,
// 				Type:    ptr.To(c.Offset(2).Lit),
// 				Kind:    paramT,
// 			})
// 			i += 2
// 			continue
// 		}
// 		// "ctx context.Context"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.IDENT &&
// 			c.PeekOffset(2).Tok == token.PERIOD &&
// 			c.PeekOffset(3).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name:    ptr.To(c.Cur().Lit),
// 				Package: c.Peek().Lit,
// 				Type:    ptr.To(c.Offset(3).Lit),
// 				Kind:    paramT,
// 			})
// 			i += 3
// 			continue
// 		}

// 		// "n int"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name: ptr.To(c.Cur().Lit),
// 				Type: ptr.To(c.Offset(1).Lit),
// 				Kind: paramT,
// 			})
// 			i += 1
// 			continue
// 		}

// 		// "v *atomic.Value"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.MUL &&
// 			c.PeekOffset(2).Tok == token.IDENT &&
// 			c.PeekOffset(3).Tok == token.PERIOD &&
// 			c.PeekOffset(4).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name:      ptr.To(c.Cur().Lit),
// 				Package:   c.PeekOffset(2).Lit,
// 				Type:      ptr.To(c.Offset(4).Lit),
// 				IsPointer: ptr.To(true),
// 				Kind:      paramT,
// 			})
// 			i += 4
// 			continue
// 		}

// 		// "n *int"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.MUL &&
// 			c.PeekOffset(2).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name:      ptr.To(c.Cur().Lit),
// 				Type:      ptr.To(c.Offset(2).Lit),
// 				IsPointer: ptr.To(true),
// 				Kind:      paramT,
// 			})
// 			i += 2
// 			continue
// 		}
// 		// "*int"
// 		if c.Cur().Tok == token.MUL &&
// 			c.PeekOffset(1).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Type:      ptr.To(c.Offset(1).Lit),
// 				IsPointer: ptr.To(true),
// 				Kind:      paramT,
// 			})
// 			i += 1
// 			continue
// 		}

// 		// "vs []*atomic.Value"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.LBRACK &&
// 			c.PeekOffset(2).Tok == token.RBRACK &&
// 			c.PeekOffset(3).Tok == token.MUL &&
// 			c.PeekOffset(4).Tok == token.IDENT &&
// 			c.PeekOffset(5).Tok == token.PERIOD &&
// 			c.PeekOffset(6).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name:      ptr.To(c.Cur().Lit),
// 				Package:   c.PeekOffset(4).Lit,
// 				Type:      ptr.To(c.Offset(6).Lit),
// 				IsPointer: ptr.To(true),
// 				IsSlice:   ptr.To(true),
// 				Kind:      paramT,
// 			})
// 			i += 6
// 			continue
// 		}
// 		// "vs []atomic.Value"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.LBRACK &&
// 			c.PeekOffset(2).Tok == token.RBRACK &&
// 			c.PeekOffset(3).Tok == token.IDENT &&
// 			c.PeekOffset(4).Tok == token.PERIOD &&
// 			c.PeekOffset(5).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name:    ptr.To(c.Cur().Lit),
// 				Package: c.PeekOffset(3).Lit,
// 				Type:    ptr.To(c.Offset(5).Lit),
// 				IsSlice: ptr.To(true),
// 				Kind:    paramT,
// 			})
// 			i += 5
// 			continue
// 		}
// 		// "[]atomic.Value"
// 		if c.Cur().Tok == token.LBRACK &&
// 			c.PeekOffset(1).Tok == token.RBRACK &&
// 			c.PeekOffset(2).Tok == token.IDENT &&
// 			c.PeekOffset(3).Tok == token.PERIOD &&
// 			c.PeekOffset(4).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Package: c.PeekOffset(2).Lit,
// 				Type:    ptr.To(c.Offset(4).Lit),
// 				IsSlice: ptr.To(true),
// 				Kind:    paramT,
// 			})
// 			i += 5
// 			continue
// 		}
// 		// "vs []*int"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.LBRACK &&
// 			c.PeekOffset(2).Tok == token.RBRACK &&
// 			c.PeekOffset(3).Tok == token.MUL &&
// 			c.PeekOffset(4).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name:      ptr.To(c.Cur().Lit),
// 				Type:      ptr.To(c.Offset(4).Lit),
// 				IsSlice:   ptr.To(true),
// 				IsPointer: ptr.To(true),
// 				Kind:      paramT,
// 			})
// 			i += 4
// 			continue
// 		}
// 		// "vs []int"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.LBRACK &&
// 			c.PeekOffset(2).Tok == token.RBRACK &&
// 			c.PeekOffset(3).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name:    ptr.To(c.Cur().Lit),
// 				Type:    ptr.To(c.Offset(3).Lit),
// 				IsSlice: ptr.To(true),
// 				Kind:    paramT,
// 			})
// 			i += 3
// 			continue
// 		}
// 		// "[]int"
// 		if c.Cur().Tok == token.LBRACK &&
// 			c.PeekOffset(1).Tok == token.RBRACK &&
// 			c.PeekOffset(2).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Type:    ptr.To(c.Offset(2).Lit),
// 				IsSlice: ptr.To(true),
// 				Kind:    paramT,
// 			})
// 			i += 2
// 			continue
// 		}
// 		// "[]*int"
// 		if c.Cur().Tok == token.LBRACK &&
// 			c.PeekOffset(1).Tok == token.RBRACK &&
// 			c.PeekOffset(2).Tok == token.MUL &&
// 			c.PeekOffset(3).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Type:      ptr.To(c.Offset(3).Lit),
// 				IsSlice:   ptr.To(true),
// 				IsPointer: ptr.To(true),
// 				Kind:      paramT,
// 			})
// 			i += 3
// 			continue
// 		}
// 		// "n ...int"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.ELLIPSIS &&
// 			c.PeekOffset(2).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name:       ptr.To(c.Cur().Lit),
// 				Type:       ptr.To(c.Offset(2).Lit),
// 				IsSlice:    ptr.To(true),
// 				IsVariadic: ptr.To(true),
// 				Kind:       paramT,
// 			})
// 			i += 2
// 			continue
// 		}
// 		// "n ...*int"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.ELLIPSIS &&
// 			c.PeekOffset(2).Tok == token.MUL &&
// 			c.PeekOffset(3).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name:       ptr.To(c.Cur().Lit),
// 				Type:       ptr.To(c.Offset(3).Lit),
// 				IsSlice:    ptr.To(true),
// 				IsVariadic: ptr.To(true),
// 				IsPointer:  ptr.To(true),
// 				Kind:       paramT,
// 			})
// 			i += 3
// 			continue
// 		}
// 		// "...int"
// 		if c.Cur().Tok == token.ELLIPSIS &&
// 			c.PeekOffset(1).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Type:       ptr.To(c.Offset(1).Lit),
// 				IsSlice:    ptr.To(true),
// 				IsVariadic: ptr.To(true),
// 				Kind:       paramT,
// 			})
// 			i += 1
// 			continue
// 		}
// 		// "...*int"
// 		if c.Cur().Tok == token.ELLIPSIS &&
// 			c.PeekOffset(1).Tok == token.MUL &&
// 			c.PeekOffset(2).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Type:       ptr.To(c.Offset(2).Lit),
// 				IsSlice:    ptr.To(true),
// 				IsVariadic: ptr.To(true),
// 				IsPointer:  ptr.To(true),
// 				Kind:       paramT,
// 			})
// 			i += 2
// 			continue
// 		}
// 		// "vs ...atomic.Value"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.ELLIPSIS &&
// 			c.PeekOffset(2).Tok == token.IDENT &&
// 			c.PeekOffset(3).Tok == token.PERIOD &&
// 			c.PeekOffset(4).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name:       ptr.To(c.Cur().Lit),
// 				Package:    c.Offset(2).Lit,
// 				Type:       ptr.To(c.Offset(4).Lit),
// 				IsSlice:    ptr.To(true),
// 				IsVariadic: ptr.To(true),
// 				Kind:       paramT,
// 			})
// 			i += 4
// 			continue
// 		}
// 		// "vs ...*atomic.Value"
// 		if c.Cur().Tok == token.IDENT &&
// 			c.PeekOffset(1).Tok == token.ELLIPSIS &&
// 			c.PeekOffset(2).Tok == token.MUL &&
// 			c.PeekOffset(3).Tok == token.IDENT &&
// 			c.PeekOffset(4).Tok == token.PERIOD &&
// 			c.PeekOffset(5).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Name:       ptr.To(c.Cur().Lit),
// 				Package:    c.Offset(3).Lit,
// 				Type:       ptr.To(c.Offset(5).Lit),
// 				IsPointer:  ptr.To(true),
// 				IsSlice:    ptr.To(true),
// 				IsVariadic: ptr.To(true),
// 				Kind:       paramT,
// 			})
// 			i += 5
// 			continue
// 		}
// 		// "...atomic.Value"
// 		if c.Cur().Tok == token.ELLIPSIS &&
// 			c.PeekOffset(1).Tok == token.IDENT &&
// 			c.PeekOffset(2).Tok == token.PERIOD &&
// 			c.PeekOffset(3).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Package:    c.Offset(1).Lit,
// 				Type:       ptr.To(c.Offset(3).Lit),
// 				IsSlice:    ptr.To(true),
// 				IsVariadic: ptr.To(true),
// 				Kind:       paramT,
// 			})
// 			i += 3
// 			continue
// 		}
// 		// "...*atomic.Value"
// 		if c.Cur().Tok == token.ELLIPSIS &&
// 			c.PeekOffset(1).Tok == token.MUL &&
// 			c.PeekOffset(2).Tok == token.IDENT &&
// 			c.PeekOffset(3).Tok == token.PERIOD &&
// 			c.PeekOffset(4).Tok == token.IDENT {
// 			lb = append(lb, &LogicBlock{
// 				Package:    c.Offset(2).Lit,
// 				Type:       ptr.To(c.Offset(4).Lit),
// 				IsPointer:  ptr.To(true),
// 				IsSlice:    ptr.To(true),
// 				IsVariadic: ptr.To(true),
// 				Kind:       paramT,
// 			})
// 			i += 4
// 			continue
// 		}

// 		// ","
// 		if c.Cur().Tok == token.COMMA {
// 			continue
// 		}
// 		// "func (...)"
// 		if c.Cur().Tok == token.FUNC {
// 			fn := extractFunc(ExtractCursor(c, TypeFunction))
// 			fn.Kind = paramT
// 			lb = append(lb, fn)
// 			continue
// 		}
// 		// "string"
// 		if c.Cur().Tok == token.IDENT && c.Cur().Lit != "" {
// 			lb = append(lb, &LogicBlock{
// 				Type: ptr.To(c.Cur().Lit),
// 				Kind: paramT,
// 			})
// 		}

// 	}

// 	return lb
// }
