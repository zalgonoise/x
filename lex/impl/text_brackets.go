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

var tokenConv = map[rune]TextToken{
	'{': TokenLBRACK,
	'}': TokenRBRACK,
}

// TemplateItem represents the lex.Item for a runes lexer based on TextToken identifiers
type TemplateItem[C TextToken, I rune] lex.Item[C, I]

// toTemplateItem converts a lex.Item type to TemplateItem
func toTemplateItem[C TextToken, T rune](i lex.Item[C, T]) TemplateItem[C, T] {
	return (TemplateItem[C, T])(i)
}

// String implements fmt.Stringer
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
	}
	return ""
}

// initState describes the StateFn to kick off the lexer / parser. It is also the default fallback StateFn
// for any other StateFn
func initState[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	for l.Cur() != 0 {
		if tokenConv[(rune)(l.Cur())] == TokenLBRACK {
			l.Prev()
			l.Emit((C)(TokenIDENT))
			l.Next()
			l.Ignore()
			return stateLBRACK[C, T, I]
		}
		if l.Cur() == '\n' || l.Pos() >= l.Len()-1 {
			l.Emit((C)(TokenIDENT))
			return nil
		}
		l.Next()
	}
	return nil
}

// stateLBRACK describes the StateFn to read the template content, emitting it as a template item
func stateLBRACK[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	if tokenConv[(rune)(l.Cur())] == TokenLBRACK {
		l.Next()
		l.Ignore()
	}
	for tokenConv[(rune)(l.Cur())] != TokenRBRACK {
		if l.Pos() >= l.Len()-1 {
			return stateError[C, T, I]
		}
		l.Next()
	}
	if tokenConv[(rune)(l.Cur())] == TokenRBRACK {
		l.Prev()
		l.Emit((C)(TokenTEMPL))
		l.Offset(2)
		l.Ignore()
		return initState[C, T, I]
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
func NewTextTmplLexer[C TextToken, T rune, I lex.Item[C, T]](input []rune) lex.Lexer[C, T, I] {
	var in = make([]T, len(input), len(input))
	for idx, i := range input {
		in[idx] = (T)(i)
	}
	return lex.NewLexer(initState[C, T, I], in)
}

// Run takes in a string `s`, processes it for templates, and returns the processed string and an error
func Run(s string) (string, error) {
	l := NewTextTmplLexer([]rune(s))
	var sb = new(strings.Builder)
	for {
		i := l.NextItem()
		if i.Typ == 0 {
			return sb.String(), nil
		}
		if i.Typ == TokenError {
			return sb.String(), fmt.Errorf("failed to parse token")
		}
		sb.WriteString(toTemplateItem(i).String())
	}
}
