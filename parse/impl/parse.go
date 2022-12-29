package impl

import (
	"github.com/zalgonoise/lex"
	"github.com/zalgonoise/x/parse"
)

func initParse[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	for t.Peek().Type != C(TokenEOF) {
		switch t.Peek().Type {
		case (C)(TokenIDENT):
			return parseText[C, T]
		case (C)(TokenLBRACE), (C)(TokenRBRACE):
			return parseTemplate[C, T]
		}
	}
	return nil
}

func parseText[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	if t.Peek().Type == (C)(TokenIDENT) {
		t.Node(t.Next())
	}
	t.Set(t.Parent())
	return initParse[C, T]
}

func parseTemplate[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	switch t.Peek().Type {
	case (C)(TokenLBRACE):
		t.Node(t.Next())
		return initParse[C, T]
	case (C)(TokenRBRACE):
		t.Node(t.Next())
		t.Set(t.Parent().Parent)
	case (C)(TokenEOF):
		t.Node(lex.NewItem[C, T](t.Cur().Pos, (C)(TokenError)))
	}
	return initParse[C, T]
}
