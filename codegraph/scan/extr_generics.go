package scan

import (
	"go/token"
)

type GenericsExtractor struct {
	f         *GoFile
	done      bool
	typeIsPkg bool
	lb        *LogicBlock
	paramIdx  int
	idx       int
}

// TODO: needs to be redone with the new design
func (e *GenericsExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {
	if e.lb.Generics == nil {
		e.lb.Generics = []*LogicBlock{&LogicBlock{
			Generics:     []*LogicBlock{},
			InputParams:  []*LogicBlock{},
			ReturnParams: []*LogicBlock{},
			BlockParams:  []*LogicBlock{},
		}}
	}
	if e.lb.Generics[e.paramIdx].Kind == 0 {
		e.lb.Generics[e.paramIdx].Kind = TypeGenericParam
	}
	switch tok {
	case token.IDENT:
		if e.lb.Generics[e.paramIdx].Name == "" {
			e.lb.Generics[e.paramIdx].Name = lit
			break
		}
		if e.lb.Generics[e.paramIdx].Type == "" {
			e.lb.Generics[e.paramIdx].Type = lit
			break
		}
		if e.typeIsPkg {
			e.lb.Generics[e.paramIdx].Package = e.lb.Generics[e.paramIdx].Package + "." + e.lb.Generics[e.paramIdx].Type
			break
		}
	// set type is actually package, not type
	case token.PERIOD:
		e.lb.Generics[e.paramIdx].Package = e.lb.Generics[e.paramIdx].Type
		e.lb.Generics[e.paramIdx].Kind = 0
		e.typeIsPkg = true
	case token.COMMA:
		e.paramIdx += 1
		e.lb.Generics = append(e.lb.Generics, &LogicBlock{})
	case token.RBRACK:
		return e.f.Type(e.idx)
	}
	return e
}
func (e *GenericsExtractor) Done() bool {
	return e.done
}
