package schedule

import "github.com/zalgonoise/lex"

func initState(l lex.Lexer[token, byte]) lex.StateFn[token, byte] {
	switch l.Next() {
	case '@':
		l.Emit(tokenAt)
		return stateException
	case '-':
		l.Emit(tokenDash)

		return initState
	case ',':
		l.Emit(tokenComma)

		return initState
	case '/':
		l.Emit(tokenSlash)

		return initState
	case '*':
		l.Emit(tokenStar)

		return initState
	case ' ':
		l.Emit(tokenSpace)

		return initState
	case 0:
		return nil
	default:
		return stateAlphanumeric
	}
}

func stateAlphanumeric(l lex.Lexer[token, byte]) lex.StateFn[token, byte] {
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

		l.Emit(tokenAlphanum)
	}

	switch l.Next() {
	case '-', ',', '/', '*', ' ':
		l.Prev()

		return initState
	default:
		l.Emit(tokenEOF)
		return nil
	}
}

func stateException(l lex.Lexer[token, byte]) lex.StateFn[token, byte] {
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

		l.Emit(tokenAlphanum)
	}

	l.Emit(tokenEOF)
	return nil
}
