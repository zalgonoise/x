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

type StructExtractor struct {
	f        *GoFile
	done     bool
	idx      int
	iidx     int
	identIdx int
	lvl      int
	parenLvl int
	braceLvl int
}

type InterfaceExtractor struct {
	f           *GoFile
	done        bool
	idx         int
	iidx        int
	paramIdx    int
	inputsDone  bool
	onInputs    bool
	returnsDone bool
	onReturns   bool
	lvl         int
	parenLvl    int
	funcParam   bool
	braceLvl    int
}

func (e *StructExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {
	return e
}

func (e *StructExtractor) Done() bool {
	return e.done
}

func (e *InterfaceExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {
	switch tok {
	case token.IDENT:
		if len(e.f.LogicBlocks[e.idx].BlockParams) == e.iidx {
			e.f.LogicBlocks[e.idx].BlockParams = append(e.f.LogicBlocks[e.idx].BlockParams, &LogicBlock{})
		}
		if e.f.LogicBlocks[e.idx].BlockParams[e.iidx].Type == "" {
			e.f.LogicBlocks[e.idx].BlockParams[e.iidx].Type = "func"
		}
		if e.f.LogicBlocks[e.idx].BlockParams[e.iidx].Name == "" {
			e.f.LogicBlocks[e.idx].BlockParams[e.iidx].Name = lit
			return e
		}

		if !e.inputsDone && e.onInputs {
			if len(e.f.LogicBlocks[e.idx].BlockParams[e.iidx].InputParams) == e.paramIdx {
				e.f.LogicBlocks[e.idx].BlockParams[e.iidx].InputParams = append(e.f.LogicBlocks[e.idx].BlockParams[e.iidx].InputParams, &LogicBlock{})
			}
			if e.f.LogicBlocks[e.idx].BlockParams[e.iidx].InputParams[e.paramIdx].Type == "" {
				e.f.LogicBlocks[e.idx].BlockParams[e.iidx].InputParams[e.paramIdx].Type = lit
			} else {
				e.f.LogicBlocks[e.idx].BlockParams[e.iidx].InputParams[e.paramIdx].Name = e.f.LogicBlocks[e.idx].BlockParams[e.iidx].InputParams[e.paramIdx].Type
				e.f.LogicBlocks[e.idx].BlockParams[e.iidx].InputParams[e.paramIdx].Type = lit
			}
		}
		if e.inputsDone && !e.returnsDone && e.onReturns {
			if len(e.f.LogicBlocks[e.idx].BlockParams[e.iidx].ReturnParams) == e.paramIdx {
				e.f.LogicBlocks[e.idx].BlockParams[e.iidx].ReturnParams = append(e.f.LogicBlocks[e.idx].BlockParams[e.iidx].ReturnParams, &LogicBlock{})
			}
			if e.f.LogicBlocks[e.idx].BlockParams[e.iidx].ReturnParams[e.paramIdx].Type == "" {
				e.f.LogicBlocks[e.idx].BlockParams[e.iidx].ReturnParams[e.paramIdx].Type = lit
			} else {
				e.f.LogicBlocks[e.idx].BlockParams[e.iidx].ReturnParams[e.paramIdx].Name = e.f.LogicBlocks[e.idx].BlockParams[e.iidx].ReturnParams[e.paramIdx].Type
				e.f.LogicBlocks[e.idx].BlockParams[e.iidx].ReturnParams[e.paramIdx].Type = lit
			}
		}

	case token.LPAREN:
		// got method func input
		if !e.inputsDone && !e.onInputs {
			e.onInputs = true
			return e
		}

		// got func within method func input;
		// or, got func on method return
		if (e.onInputs && !e.inputsDone) || (e.inputsDone && e.funcParam) {
			e.parenLvl += 1
			return e
		}

		if e.inputsDone && !e.funcParam && !e.returnsDone {
			e.onReturns = true
			return e
		}
	case token.RPAREN:
		if e.onInputs && !e.inputsDone && e.parenLvl == 0 {
			e.inputsDone = true
			e.paramIdx = 0
			return e
		}
		if e.onReturns && e.returnsDone && e.parenLvl == 0 {
			e.returnsDone = true
			e.paramIdx = 0
			return e
		}

		if e.parenLvl > 0 {
			e.parenLvl -= 1
			if e.parenLvl == 0 {
				e.funcParam = false
			}
			return e
		}
	case token.FUNC:
		e.funcParam = true
		return e
	case token.COMMA:
		e.paramIdx += 1
		return e

	case token.SEMICOLON:
		if e.parenLvl > 0 || e.braceLvl > 0 {
			return e
		}

		e.returnsDone = false
		e.inputsDone = false
		if len(e.f.LogicBlocks[e.idx].BlockParams[e.iidx].InputParams) == 0 && len(e.f.LogicBlocks[e.idx].BlockParams[e.iidx].ReturnParams) == 0 {
			e.f.LogicBlocks[e.idx].BlockParams[e.iidx].Type = e.f.LogicBlocks[e.idx].BlockParams[e.iidx].Name
		}
		if e.f.LogicBlocks[e.idx].BlockParams[e.iidx].Kind == 0 {
			e.f.LogicBlocks[e.idx].BlockParams[e.iidx].Kind = TypeMethod
		}
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

func (e *InterfaceExtractor) Done() bool {
	return e.done
}

func (e *ElementsExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {
	if len(e.f.LogicBlocks[e.idx].BlockParams) == 0 {
		e.f.LogicBlocks[e.idx].BlockParams = append(e.f.LogicBlocks[e.idx].BlockParams, &LogicBlock{})
	}

	switch e.f.LogicBlocks[e.idx].Kind {
	case TypeInterface:
		return e.f.Interface(e.idx, e.lvl).Do(pos, tok, lit)
	case TypeStruct:
		return e.f.Struct(e.idx, e.lvl).Do(pos, tok, lit)
	default:
		return e
	}
}
func (e *ElementsExtractor) Done() bool {
	return e.done
}
