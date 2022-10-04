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
		e.f.SetITFType(e.idx, e.iidx)
		e.f.SetITFName(e.idx, e.iidx, lit)

		if !e.inputsDone && e.onInputs {
			e.f.LBlock(e,
				NewFilter("logicBlock", e.idx),
				NewFilter("input", e.paramIdx),
			).Do(pos, tok, lit)
		}
		if e.inputsDone && !e.returnsDone && e.onReturns {
			e.f.LBlock(e,
				NewFilter("logicBlock", e.idx),
				NewFilter("return", e.paramIdx),
			).Do(pos, tok, lit)
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
		if !e.inputsDone && e.onInputs {
			e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).InputParam(e.paramIdx).IsFunc()
		}
		if e.inputsDone && !e.returnsDone && e.onReturns {
			e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).ReturnParam(e.paramIdx).IsFunc()
		}

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
		if e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).InputLen() == 0 && e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).ReturnLen() == 0 {
			e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).SetType(e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).Name)
		}
		if e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).Kind == 0 {
			e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).SetKind(TypeMethod)
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
