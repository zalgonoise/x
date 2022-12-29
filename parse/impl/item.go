package impl

import (
	"fmt"

	"github.com/zalgonoise/lex"
)

// TextToken is a unique identifier for this text template implementation
type TextToken int

const (
	TokenEOF TextToken = iota
	TokenError
	TokenIDENT
	TokenLBRACE
	TokenRBRACE

	TokenRoot
	TokenText
	TokenTemplate
	TokenTemplateEnd
)

// TemplateItem represents the lex.Item for a runes lexer based on TextToken identifiers
type TemplateItem[C TextToken, I rune] lex.Item[C, I]

// toTemplateItem converts a lex.Item type to TemplateItem
func toTemplateItem[C TextToken, T rune](i lex.Item[C, T]) TemplateItem[C, T] {
	return (TemplateItem[C, T])(i)
}

// String implements fmt.Stringer; which is processing each TemplateItem as a string
func (t TemplateItem[C, I]) String() string {
	switch t.Type {
	case C(TokenIDENT):
		var rs = make([]rune, len(t.Value), len(t.Value))
		for idx, r := range t.Value {
			rs[idx] = (rune)(r)
		}
		return string(rs)
	case C(TokenLBRACE):
		return ">>"
	case C(TokenRBRACE):
		return "<<"
	case C(TokenError):
		return fmt.Sprintf("[error on line: %d]", t.Pos)
	case C(TokenEOF):
		return "" // placeholder action for EOF tokens
	}
	return ""
}
