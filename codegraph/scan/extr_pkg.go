package scan

import "go/token"

type PackageExtractor struct {
	f    *GoFile
	done bool
}

func (e *PackageExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor {
	switch tok {
	case token.IDENT:
		e.f.PackageName = lit
	case token.SEMICOLON:
		e.done = true
	}
	return e
}
func (e *PackageExtractor) Done() bool {
	return e.done
}
