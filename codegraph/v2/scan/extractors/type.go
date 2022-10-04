package scan

import (
	"go/token"
)

type TypeExtractor struct {
	f    *GoFile
	done bool
	idx  int
	lvl  int
}

func (e *TypeExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {

	switch tok {
	case token.IDENT:
		e.f.SetLBName(e.idx, lit)
	case token.LBRACK:
		return e.f.Generics(e.idx)
	case token.INTERFACE:
		e.f.SetLBKind(e.idx, TypeInterface)
		return e.f.Interface(e, e.idx, e.lvl)
	case token.STRUCT:
		e.f.SetLBKind(e.idx, TypeStruct)
		return e.f.Struct(e, e.idx, e.lvl)
	// case token.LBRACE:
	// e.lvl += 1
	// return e.f.Element(e.idx, e.lvl)
	case token.RBRACE:
		// if e.lvl > 0 {
		// 	e.lvl -= 1
		// 	return e
		// }
		e.done = true
	}
	return e
}
func (e *TypeExtractor) Done() bool {
	return e.done
}
