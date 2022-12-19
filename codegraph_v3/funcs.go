package codegraph

import (
	"go/token"
)

func (wt *WithTokens) Func() error {
	if wt.LogicBlocks == nil {
		wt.LogicBlocks = []*LogicBlock{}
	}
	count := len(wt.LogicBlocks)

	for i := wt.Tokens.Pos(); i < wt.Tokens.Len(); i++ {
		if wt.Tokens.Next().Tok == token.FUNC {
			err := extractFunc(wt, count)
			if err != nil {
				return err
			}
			count++
		}
	}
	return nil
}
func extractFunc(wt *WithTokens, count int) error {
	fn := &LogicBlock{
		Kind: TypeFunction,
	}
	if wt.Tokens.Cur().Tok == token.FUNC &&
		wt.Tokens.Peek().Tok == token.IDENT {
		fn.Name = wt.Tokens.Next().Lit
	}

	switch wt.Tokens.Next().Tok {
	// TODO: add generics handler
	case token.LPAREN:
		extractInput(wt, fn)
	}
	if wt.GoFile.LogicBlocks == nil {
		wt.GoFile.LogicBlocks = []*LogicBlock{}
	}
	wt.GoFile.LogicBlocks = append(wt.GoFile.LogicBlocks, fn)

	return nil
}

func extractInput(wt *WithTokens, lb *LogicBlock) {
	var (
		parenLvl = 0
	)
	if lb.InputParams == nil {
		lb.InputParams = []*Identifier{}
	}

	if wt.Tokens.Cur().Tok != token.LPAREN {
		return // not an input parameter
	}
	parenLvl++

	for wt.Tokens.Next().Tok != token.RPAREN {
		switch wt.Tokens.Cur().Tok {
		case token.IDENT:
			if wt.Tokens.PeekOffset(1).Tok == token.PERIOD &&
				wt.Tokens.PeekOffset(2).Tok == token.IDENT {
				lb.InputParams = append(lb.InputParams, &Identifier{
					Package: wt.Tokens.Cur().Lit,
					Type:    wt.Tokens.PeekOffset(2).Lit,
				})
				wt.Tokens.Offset(2)
				continue
			}
			if wt.Tokens.PeekOffset(1).Tok == token.IDENT &&
				wt.Tokens.PeekOffset(2).Tok == token.PERIOD &&
				wt.Tokens.PeekOffset(3).Tok == token.IDENT {
				lb.InputParams = append(lb.InputParams, &Identifier{
					Name:    wt.Tokens.Cur().Lit,
					Package: wt.Tokens.Peek().Lit,
					Type:    wt.Tokens.PeekOffset(3).Lit,
				})
				wt.Tokens.Offset(3)
				continue
			}
			if wt.Tokens.PeekOffset(1).Tok == token.IDENT {
				lb.InputParams = append(lb.InputParams, &Identifier{
					Name: wt.Tokens.Cur().Lit,
					Type: wt.Tokens.PeekOffset(1).Lit,
				})
				wt.Tokens.Offset(1)
				continue
			}
			lb.InputParams = append(lb.InputParams, &Identifier{
				Type: wt.Tokens.Cur().Lit,
			})
		}
	}
}
