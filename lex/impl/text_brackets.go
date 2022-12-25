package impl

import (
	"fmt"
	"strings"

	"github.com/zalgonoise/x/lex"
)

// TextToken is a unique identifier for this text template implementation
type TextToken int

const (
	TokenEOF TextToken = iota
	TokenError
	TokenIDENT
	TokenTEMPL
	TokenLBRACK
	TokenRBRACK
)

// TemplateItem represents the lex.Item for a runes lexer based on TextToken identifiers
type TemplateItem[C TextToken, I rune] lex.Item[C, I]

// toTemplateItem converts a lex.Item type to TemplateItem
func toTemplateItem[C TextToken, T rune](i lex.Item[C, T]) TemplateItem[C, T] {
	return (TemplateItem[C, T])(i)
}

// String implements fmt.Stringer; which is processing each TemplateItem as a string
func (t TemplateItem[C, I]) String() string {
	switch t.Typ {
	case C(TokenIDENT):
		var rs = make([]rune, len(t.Val), len(t.Val))
		for idx, r := range t.Val {
			rs[idx] = (rune)(r)
		}
		return string(rs)
	case C(TokenTEMPL):
		var rs = make([]rune, len(t.Val), len(t.Val))
		for idx, r := range t.Val {
			rs[idx] = (rune)(r)
		}
		return ">>" + string(rs) + "<<"
	case C(TokenError):
		return ":ERR:"
	case C(TokenEOF):
		return "" // placeholder action for EOF tokens
	}
	return ""
}

// initState describes the StateFn to kick off the lexer / parser. It is also the default fallback StateFn
// for any other StateFn
func initState[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	for l.Peek() != 0 {
		// accept all runes that are not '{' or EOF
		l.AcceptRun(func(item T) bool {
			return item != '{' && item != 0
		})
		// if there is more than one accepted token, emit it/them
		if l.Width() > 0 {
			l.Emit((C)(TokenIDENT))
			l.Ignore()
		}
		// peek into the next rune
		switch l.Peek() {
		case '{':
			// advance and process a template string
			l.Next()
			return stateLBRACK[C, T, I]
		case 0:
			// EOF ; break the loop
			break
		default:
			// no more actions; move to the next rune
			l.Next()
		}
	}

	// reached eof; emit tokens if word width is not zero
	if l.Width() > 0 {
		l.Emit((C)(TokenIDENT))
		l.Ignore()
	}
	// emit EOF token for reference (when parsing)
	l.Emit((C)(TokenEOF))
	l.Ignore()
	return nil
}

// stateLBRACK describes the StateFn to read the template content, emitting it as a template item
func stateLBRACK[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	// advance the cursor if we're at the LBRACK token
	if l.Cur() == '{' {
		l.Next()
		l.Ignore()
	}
	// accept all character until it hits a RBRACK or EOF
	l.AcceptRun(func(item T) bool {
		return item != '}' && item != 0
	})

	// peek into the next rune
	switch l.Peek() {
	case '}':
		// reached end of template, emit template and return to initState func
		l.Emit((C)(TokenTEMPL))
		l.Offset(2) // advance to the char after '}'
		l.Ignore()
		return initState[C, T, I]
	case 0:
		// template was not closed, return an error
		return stateError[C, T, I]
	}
	return nil
}

// stateError describes an errored state in the lexer / parser, ignoring this set of tokens and emitting an
// error item
func stateError[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	l.Emit((C)(TokenError))
	l.Ignore()
	return initState[C, T, I]
}

// NewTextTmplLexer creates a text template lexer based on the input slice of runes
func NewTextTmplLexer[C TextToken, T rune, I lex.Item[C, T]](input []T) lex.Lexer[C, T, I] {
	return lex.NewLexer(initState[C, T, I], input)
}

// Run takes in a string `s`, processes it for templates, and returns the processed string and an error
func Run(s string) (string, error) {
	l := NewTextTmplLexer([]rune(s))
	var sb = new(strings.Builder)
	for {
		i := l.NextItem()
		sb.WriteString(toTemplateItem(i).String())

		if i.Typ == 0 {
			return sb.String(), nil
		}
		if i.Typ == TokenError {
			return sb.String(), fmt.Errorf("failed to parse token")
		}
	}
}
