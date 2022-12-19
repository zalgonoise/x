package codegraph

import (
	"go/token"

	cur "github.com/zalgonoise/cur"
)

type ExtractorTarget struct{}

var (
	FuncExtractor     ExtractorTarget = struct{}{}
	ParamExtractor    ExtractorTarget = struct{}{}
	GenericsExtractor ExtractorTarget = struct{}{}
)

type extractorFilter struct {
	init   token.Token
	sep    token.Token
	closer token.Token
}

var filterMap = map[LogicBlockKind]*extractorFilter{
	TypeFunction: {
		init:   token.FUNC,
		sep:    token.FUNC,
		closer: token.LBRACE,
	},
	TypeFuncParam: {
		init:   token.LPAREN,
		sep:    token.LPAREN,
		closer: token.RPAREN,
	},
	TypeGenericParam: {
		init:   token.LBRACK,
		sep:    token.LBRACK,
		closer: token.RBRACK,
	},
}

func ExtractCursor(c cur.Cursor[GoToken], target LogicBlockKind) cur.Cursor[GoToken] {
	return extractCursor(c, filterMap[target])
}

func extractCursor(c cur.Cursor[GoToken], filter *extractorFilter) cur.Cursor[GoToken] {
	if filter == nil {
		return nil
	}

	if c.Cur().Tok != filter.init {
		return nil
	}

	var (
		lvl   int = 0
		start int = c.Pos()
		end   int
	)

	for i := start; i < c.Len(); i++ {
		c.Next()
		switch c.Cur().Tok {
		case filter.sep:
			lvl++
		case filter.closer:
			if lvl > 0 {
				lvl--
			}
			if lvl == 0 {
				end = c.Pos()
				slice := c.Extract(start, end)
				return cur.NewCursor(slice)
			}
		}
	}
	return nil
}
