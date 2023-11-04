package cronlex

import (
	"github.com/zalgonoise/lex"
)

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
		l.Emit(TokenEOF)

		return nil
	default:
		return stateAlphanumeric
	}
}

func stateAlphanumeric(l lex.Lexer[Token, byte]) lex.StateFn[Token, byte] {
	l.Backup() // undo l.Next() for the l.AcceptRun call

	for {
		if item := l.Cur(); (item >= '0' && item <= '9') || (item >= 'A' && item <= 'Z') || (item >= 'a' && item <= 'z') {
			l.Next()

			continue
		}
		break
	}

	if l.Width() > 0 {
		l.Emit(TokenAlphaNum)
	}

	return StateFunc
}

func stateException(l lex.Lexer[Token, byte]) lex.StateFn[Token, byte] {
	l.Backup() // undo l.Next() for the l.AcceptRun call

	for {
		if item := l.Cur(); (item >= 'A' && item <= 'Z') || (item >= 'a' && item <= 'z') {
			l.Next()

			continue
		}
		break
	}

	if l.Width() > 0 {
		l.Emit(TokenAlphaNum)
	}

	return StateFunc
}
