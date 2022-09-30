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
	if len(e.f.LogicBlocks) == e.idx {
		e.f.LogicBlocks = append(e.f.LogicBlocks, &LogicBlock{
			Generics:     []*LogicBlock{},
			InputParams:  []*LogicBlock{},
			ReturnParams: []*LogicBlock{},
			BlockParams:  []*LogicBlock{},
		})
	}

	switch tok {
	case token.IDENT:
		if e.f.LogicBlocks[e.idx].Name == "" {
			e.f.LogicBlocks[e.idx].Name = lit
		}
	case token.LBRACK:
		return e.f.Generics(e.idx)
	case token.INTERFACE:
		if e.f.LogicBlocks[e.idx].Kind == 0 {
			e.f.LogicBlocks[e.idx].Kind = TypeInterface
		}
	case token.STRUCT:
		if e.f.LogicBlocks[e.idx].Kind == 0 {
			e.f.LogicBlocks[e.idx].Kind = TypeStruct
		}
	case token.LBRACE:
		e.lvl += 1
		return e.f.Element(e.idx, e.lvl)
	case token.RBRACE:
		if e.lvl > 0 {
			e.lvl -= 1
			return e
		}
		e.done = true
	}
	return e
}
func (e *TypeExtractor) Done() bool {
	return e.done
}
