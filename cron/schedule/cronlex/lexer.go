package cronlex

import "github.com/zalgonoise/lex"

func StateFunc(l lex.Lexer[Token, byte]) lex.StateFn[Token, byte] {
	switch l.Next() {
	case '@':
		l.Emit(TokenAt)
		return stateException
	case '-':
		l.Emit(TokenDash)

		return StateFunc
	case ',':
		l.Emit(TokenComma)

		return StateFunc
	case '/':
		l.Emit(TokenSlash)

		return StateFunc
	case '*':
		l.Emit(TokenStar)

		return StateFunc
	case ' ':
		l.Emit(TokenSpace)

		return StateFunc
	case 0:
		return nil
	default:
		return stateAlphanumeric
	}
}

func stateAlphanumeric(l lex.Lexer[Token, byte]) lex.StateFn[Token, byte] {
	l.AcceptRun(func(item byte) bool {
		switch {
		case item >= '0' && item <= '9', item >= 'A' && item <= 'Z', item >= 'a' && item <= 'z':
			return true
		default:
			return false
		}
	})

	if l.Width() > 0 {
		// advance cursor on last item
		if item := l.Peek(); item == 0 {
			l.Next()
		}

		l.Emit(TokenAlphaNum)
	}

	switch l.Next() {
	case '-', ',', '/', '*', ' ':
		l.Prev()

		return StateFunc
	default:
		l.Emit(TokenEOF)
		return nil
	}
}

func stateException(l lex.Lexer[Token, byte]) lex.StateFn[Token, byte] {
	l.AcceptRun(func(item byte) bool {
		switch {
		case item >= 'A' && item <= 'Z', item >= 'a' && item <= 'z':
			return true
		default:
			return false
		}
	})

	if l.Width() > 0 {
		// advance cursor on last item
		if item := l.Peek(); item == 0 {
			l.Next()
		}

		l.Emit(TokenAlphaNum)
	}

	l.Emit(TokenEOF)
	return nil
}
