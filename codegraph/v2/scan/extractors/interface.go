package scan

import (
	"go/token"
)

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
	braceLvl    int
	parent      Extractor
}

func (e *InterfaceExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {
	switch tok {
	case token.LBRACE:
		e.braceLvl += 1
	case token.RBRACE:
		if e.braceLvl > 0 {
			e.braceLvl -= 1
		}
		if e.braceLvl == 0 {
			e.done = true
			return e.parent.Do(pos, tok, lit)
		}

	case token.IDENT:
		e.f.SetITFType(e.idx, e.iidx)
		e.f.SetITFName(e.idx, e.iidx, lit)

	case token.LPAREN:
		e.parenLvl += 1
		// got method func input
		if !e.inputsDone && !e.onInputs {
			e.onInputs = true
			return e.f.BlockParam(e, e.idx, e.iidx, TargetInput).Do(pos, tok, lit)
		}

		if e.inputsDone && !e.returnsDone {
			e.onReturns = true
			return e.f.BlockParam(e, e.idx, e.iidx, TargetReturn).Do(pos, tok, lit)
		}
	case token.RPAREN:
		if e.onInputs && !e.inputsDone && e.parenLvl == 0 {
			e.inputsDone = true
			e.paramIdx = 0
			e.parenLvl -= 1
			return e
		}
		if e.onReturns && e.returnsDone && e.parenLvl == 0 {
			e.returnsDone = true
			e.paramIdx = 0
			e.parenLvl -= 1
			return e
		}

		if e.parenLvl > 0 {
			e.parenLvl -= 1
			return e
		}

	case token.SEMICOLON:
		e.iidx += 1
		if e.parenLvl > 0 || e.braceLvl > 0 {
			return e
		}

		e.returnsDone = false
		e.inputsDone = false
		e.onInputs = false
		e.onReturns = false
		if e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).InputLen() == 0 && e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).ReturnLen() == 0 {
			e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).SetType(e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).Name)
		}
		if e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).Kind == 0 {
			e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).SetKind(TypeMethod)
		}
		e.iidx += 1
	}
	return e
}

func (e *InterfaceExtractor) Done() bool {
	return e.done
}
