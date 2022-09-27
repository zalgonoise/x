package scan

import "go/token"

type ElementsExtractor struct {
	f    *GoFile
	done bool
	// extDone bool
	lb       *LogicBlock
	idx      int
	identIdx int
}

func (e *ElementsExtractor) Do(tok token.Token, lit string) Extractor {
	if e.lb.BlockParams == nil {
		e.lb.BlockParams = []*Param{}
	}
	if e.lb.BlockParams[e.idx] == nil {
		e.lb.BlockParams[e.idx] = &Param{}
	}

	switch tok {
	case token.IDENT:
		switch e.identIdx {
		case 0:
			e.lb.BlockParams[e.idx].Name = lit
			e.identIdx += 1
		case 1:
			e.lb.BlockParams[e.idx].Type = lit
		}

	case token.LPAREN:
		if e.lb.Type == TypeInterface {
			e.lb.BlockParams[e.idx].Type = "METHOD"
		}
	case token.SEMICOLON:
		e.identIdx = 0
		e.idx += 1
	case token.RBRACE:
		e.done = true
	}
	return e
}
func (e *ElementsExtractor) Done() bool {
	return e.done
}
