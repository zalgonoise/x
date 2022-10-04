package scan

import "go/token"

type LogicBlockExtractor struct {
	parent Extractor
	lb     *LogicBlock
	e      Extractor
	done   bool
}

func (l *LogicBlockExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {
	switch tok {
	case token.IDENT:
		l.lb.SetType(lit)
	case token.SEMICOLON:
		l.e = l.parent
		l.done = true
	}

	return l.e
}

func (l *LogicBlockExtractor) Done() bool {
	return l.done
}
