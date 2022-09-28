package scan

import (
	"go/token"
	"strings"
)

type ImportExtractor struct {
	f    *GoFile
	done bool
}

func (e *ImportExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {
	switch tok {
	case token.LPAREN:
		e.done = false
	case token.STRING:
		imp := e.proc(lit)
		e.f.Imports = append(e.f.Imports, imp)
	case token.RPAREN:
		e.done = true
	}
	return e
}
func (e *ImportExtractor) Done() bool {
	return e.done
}

func (e *ImportExtractor) proc(lit string) *Import {
	repl := strings.ReplaceAll(lit, `"`, "")
	s := strings.Split(repl, "/")
	return &Import{
		Package: s[len(s)-1],
		URI:     repl,
	}
}
