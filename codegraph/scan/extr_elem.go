package scan

import (
	"go/token"
)

type ElementsExtractor struct {
	f        *GoFile
	done     bool
	idx      int
	iidx     int
	identIdx int
	lvl      int
	parenLvl int
	braceLvl int
}

func (e *ElementsExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {
	if len(e.f.LogicBlocks[e.idx].BlockParams) == 0 {
		e.f.LogicBlocks[e.idx].BlockParams = append(e.f.LogicBlocks[e.idx].BlockParams, &LogicBlock{})
	}

	switch tok {
	case token.IDENT:
		if len(e.f.LogicBlocks[e.idx].BlockParams) == e.iidx {
			e.f.LogicBlocks[e.idx].BlockParams = append(e.f.LogicBlocks[e.idx].BlockParams, &LogicBlock{})
		}
		if e.f.LogicBlocks[e.idx].BlockParams[e.iidx].Name == "" {
			e.f.LogicBlocks[e.idx].BlockParams[e.iidx].Name = lit
		}

	case token.LPAREN:
		if e.f.LogicBlocks[e.idx].Kind == TypeInterface {
			e.f.LogicBlocks[e.idx].BlockParams[e.iidx].Kind = TypeInterface
		}
		e.parenLvl += 1
	case token.RPAREN:
		if e.parenLvl > 0 {
			e.parenLvl -= 1
			return e
		}
	case token.SEMICOLON:
		if e.parenLvl > 0 || e.braceLvl > 0 {
			return e
		}

		e.identIdx = 0
		e.iidx += 1
	case token.LBRACE:
		e.braceLvl += 1
	case token.RBRACE:
		if e.braceLvl > 0 {
			e.braceLvl -= 1
			return e
		}
		e.done = true
		return e
	}
	return e
}
func (e *ElementsExtractor) Done() bool {
	return e.done
}
