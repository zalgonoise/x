package impl

import (
	"github.com/zalgonoise/lex"
	"github.com/zalgonoise/x/parse"
)

func TextTemplateParser[C TextToken, T rune](l lex.Lexer[C, T, lex.Item[C, T]]) *parse.Tree[C, T] {
	return parse.New(l, initParse[C, T], (C)(TokenRoot))
}

func initParse[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {

	return nil
}

func processFn[C TextToken, T rune, R string](t *parse.Tree[C, T]) R {

	return ""
}
