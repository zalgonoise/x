package impl

import (
	"github.com/zalgonoise/lex"
)

// initState describes the StateFn to kick off the lexer. It is also the default fallback StateFn
// for any other StateFn
func initState[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	switch l.Next() {
	case '}':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenIDENT))
		}
		l.Ignore()
		return stateRBRACE[C, T, I]
	case '{':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenIDENT))
		}
		l.Ignore()
		return stateLBRACE[C, T, I]
	case 0:
		return nil
	default:
		return stateIDENT[C, T, I]
	}
}

// stateIDENT describes the StateFn to parse text tokens.
func stateIDENT[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	l.AcceptRun(func(item T) bool {
		return item != '}' && item != '{' && item != 0
	})
	switch l.Next() {
	case '}':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenIDENT))
		}
		// l.Ignore()
		return stateRBRACE[C, T, I]
	case '{':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenIDENT))
		}
		// l.Ignore()
		return stateLBRACE[C, T, I]
	default:
		if l.Width() > 0 {
			l.Emit((C)(TokenIDENT))
		}
		l.Emit((C)(TokenEOF))
		return nil
	}
}

// stateLBRACE describes the StateFn to check for and emit an LBRACE token
func stateLBRACE[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	if l.Check(func(item T) bool {
		return item == '{'
	}) {
		l.Next() // skip this symbol
		l.Emit((C)(TokenLBRACE))
		return stateIDENT[C, T, I]
	}

	return stateError[C, T, I]

}

// stateRBRACE describes the StateFn to check for and emit an RBRACE token
func stateRBRACE[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	if l.Check(func(item T) bool {
		return item == '}'
	}) {
		l.Next() // skip this symbol
		l.Emit((C)(TokenRBRACE))
		return stateIDENT[C, T, I]
	}

	return stateError[C, T, I]

}

// stateError describes an errored state in the lexer / parser, ignoring this set of tokens and emitting an
// error item
func stateError[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	l.Backup()
	l.Prev() // mark the opening bracket as erroring token
	l.Emit((C)(TokenError))
	return initState[C, T, I]
}

// TextTemplateLexer creates a text template lexer based on the input slice of runes
func TextTemplateLexer[C TextToken, T rune, I lex.Item[C, T]](input []T) lex.Lexer[C, T, I] {
	return lex.New(initState[C, T, I], input)
}

// Run takes in a string `s`, processes it for templates, and returns the processed string and an error
func Run(s string) (string, error) {
	l := TextTemplateLexer([]rune(s))
	p := TextTemplateParser(l)
	p.Parse()
	return processFn(p), nil
	// var sb = new(strings.Builder)
	// for {
	// 	i := l.NextItem()
	// 	sb.WriteString(toTemplateItem(i).String())

	// 	switch i.Type {
	// 	case 0:
	// 		return sb.String(), nil
	// 	case TokenError:
	// 		return sb.String(), fmt.Errorf("failed to parse token")
	// 	}
	// }
}
