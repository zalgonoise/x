package scan

import (
	"fmt"
	"go/token"
)

type GenericsExtractor struct {
	f         *GoFile
	done      bool
	typeIsPkg bool
	lb        *LogicBlock
	paramIdx  int
}

func (e *GenericsExtractor) Do(tok token.Token, lit string) Extractor {
	if e.lb.Generics == nil {
		e.lb.Generics = []*Param{{}}
	}
	switch tok {
	case token.IDENT:
		fmt.Println(lit)
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
		e.lb.Generics[e.paramIdx].Type = ""
		e.typeIsPkg = true
	case token.COMMA:
		e.paramIdx += 1
		e.lb.Generics = append(e.lb.Generics, &Param{})
	case token.RBRACK:
		return e.f.Type()
	}
	return e
}
func (e *GenericsExtractor) Done() bool {
	return e.done
}
