package scan

import "go/token"

type StructExtractor struct {
	f    *GoFile
	done bool
	idx  int
	lvl  int
	// iidx     int
	// identIdx int
	// parenLvl int
	// braceLvl int
	parent Extractor
}

func (e *StructExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {
	return e.parent
}

func (e *StructExtractor) Done() bool {
	return e.done
}
