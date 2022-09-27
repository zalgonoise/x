package scan

import "go/token"

type TypeExtractor struct {
	f    *GoFile
	done bool
	idx  int
	lvl  int
}

func (e *TypeExtractor) Do(tok token.Token, lit string) Extractor {
	if len(e.f.LogicBlocks) == (e.idx) {
		e.f.LogicBlocks = append(e.f.LogicBlocks, &LogicBlock{
			Generics:     []*Param{},
			InputParams:  []*Param{},
			ReturnParams: []*Param{},
			BlockParams:  []*Param{},
		})
	}
	// temp exit out
	// if e.extDone {
	// 	e.done = true
	// 	f.LogicBlocks = append(f.LogicBlocks, e.lb)
	// 	return e
	// }
	switch tok {
	case token.IDENT:
		e.f.LogicBlocks[e.idx].Name = lit
	case token.LBRACK:
		return e.f.Generics()
	case token.INTERFACE:
		e.f.LogicBlocks[e.idx].Type = TypeInterface
	case token.STRUCT:
		e.f.LogicBlocks[e.idx].Type = TypeStruct
	case token.LBRACE:
		e.lvl += 1
		return e.f.Element()
	}
	return e
}
func (e *TypeExtractor) Done() bool {
	return e.done
}
