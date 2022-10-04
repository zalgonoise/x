package scan

import (
	"go/token"
)

type ParamExtractor struct {
	f        *GoFile
	done     bool
	idx      int
	iidx     int
	target   Target
	parent   Extractor
	parenLvl int
	paramIdx int
}

func (e *ParamExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {
	switch tok {
	case token.LPAREN:
		e.parenLvl += 1
	case token.RPAREN:
		if e.parenLvl > 1 {
			e.parenLvl -= 1
			return e
		}
		return e.parent.Do(pos, tok, lit)
	case token.IDENT:
		e.setType(lit)
	case token.PERIOD:
		e.setType(lit)
	case token.MUL:
		e.setType(lit)
	case token.COMMA:
		e.paramIdx += 1
		return e
	case token.FUNC:
		e.setType("func")
		// TODO:return func extractor
		// return e.f.BlockParam(e, e.idx, e.iidx, TargetInput)
	}

	return e
}

func (e *ParamExtractor) Done() bool {
	return e.done
}

func (e *ParamExtractor) setType(lit string) {
	switch e.target {
	case TargetReceiver:
		e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).Receiver().SetType(lit)
	case TargetInput:
		e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).InputParam(e.paramIdx).SetType(lit)
	case TargetReturn:
		e.f.GetLogicBlock(e.idx).BlockParam(e.iidx).ReturnParam(e.paramIdx).SetType(lit)
	}
}
