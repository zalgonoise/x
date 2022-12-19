package codegraph

import (
	"go/token"

	"github.com/zalgonoise/x/ptr"
)

func (w *WithTokens) Package() error {
	if w.Tokens.Cur().Tok == token.PACKAGE && w.Tokens.Peek().Tok == token.IDENT {
		packageTok := w.Tokens.Next()
		w.GoFile.PackageName = packageTok.Lit
		if w.GoFile.PackageName == "main" {
			w.GoFile.IsMain = ptr.To(true)
		}
		return nil
	}
	return NotFound(token.PACKAGE)
}
