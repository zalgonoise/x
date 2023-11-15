package cronlex

import (
	"github.com/zalgonoise/parse"
)

func ParseFunc(t *parse.Tree[Token, byte]) parse.ParseFn[Token, byte] {
	switch t.Peek().Type {
	case TokenAt:
		return parseAt
	case TokenStar:
		return parseStar
	case TokenAlphaNum:
		return parseAlphanum
	case TokenEOF:
		return nil
	default:
		return nil
	}
}

func parseAt(t *parse.Tree[Token, byte]) parse.ParseFn[Token, byte] {
	t.Node(t.Next())

	switch t.Peek().Type {
	case TokenAlphaNum:
		return parseAlphanum
	default:
		item := t.Next()
		item.Type = TokenError
		_ = t.Set(t.Parent())

		return ParseFunc
	}
}

func parseStar(t *parse.Tree[Token, byte]) parse.ParseFn[Token, byte] {
	t.Node(t.Next())

	switch t.Peek().Type {
	case TokenSpace:
		_ = t.Set(t.Parent())
		t.Next()

		return ParseFunc
	case TokenSlash:
		return parseAlphanum
	default:
		_ = t.Set(t.Parent())

		return nil
	}
}

func parseAlphanumSymbols(t *parse.Tree[Token, byte]) parse.ParseFn[Token, byte] {
	t.Node(t.Next())

	switch t.Peek().Type {
	case TokenAlphaNum:
		t.Node(t.Next())
		_ = t.Set(t.Parent().Parent)

		return parseAlphanum
	default:
		item := t.Next()
		item.Type = TokenError
		t.Node(item)

		return ParseFunc
	}
}

func parseAlphanum(t *parse.Tree[Token, byte]) parse.ParseFn[Token, byte] {
	switch t.Peek().Type {
	case TokenAlphaNum:
		t.Node(t.Next())

		return parseAlphanum
	case TokenComma, TokenDash, TokenSlash:
		return parseAlphanumSymbols
	case TokenSpace:
		_ = t.Set(t.Parent())
		t.Next()

		return ParseFunc
	}

	return nil
}
