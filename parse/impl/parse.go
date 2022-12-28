package impl

import (
	"github.com/zalgonoise/lex"
	"github.com/zalgonoise/x/parse"
)

func TextTemplateParser[T TextToken, V rune]() (*parse.Tree[T, V], chan lex.Item[T, V]) {
	return parse.New(initParse[T, V], (T)(TokenRoot))
}

func initParse[T TextToken, V rune](t *parse.Tree[T, V]) parse.ParseFn[T, V] {

	return nil
}
