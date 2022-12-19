package codegraph

import (
	"go/token"
)

func (w *WithTokens) Package() error {
	if w.Tokens.Cur().Tok == token.PACKAGE && w.Tokens.Peek().Tok == token.IDENT {
		packageTok := w.Tokens.Next()
		w.GoFile.PackageName = packageTok.Lit
		if w.GoFile.PackageName == "main" {
			w.GoFile.IsMain = true
		}
		return nil
	}
	return NotFound(token.PACKAGE)
}
