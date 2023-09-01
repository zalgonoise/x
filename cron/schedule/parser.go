package schedule

import (
	"github.com/zalgonoise/parse"
)

func initParse(t *parse.Tree[token, byte]) parse.ParseFn[token, byte] {
	switch t.Peek().Type {
	case tokenAt:
		return parseAt
	case tokenStar:
		return parseStar
	case tokenAlphanum:
		return parseAlphanum
	case tokenEOF:
		return nil
	default:
		return nil
	}
}

func parseAt(t *parse.Tree[token, byte]) parse.ParseFn[token, byte] {
	t.Node(t.Next())

	switch t.Peek().Type {
	case tokenAlphanum:
		return parseAlphanum
	default:
		item := t.Next()
		item.Type = tokenError
		t.Set(t.Parent())

		return initParse
	}
}

func parseStar(t *parse.Tree[token, byte]) parse.ParseFn[token, byte] {
	t.Node(t.Next())

	switch t.Peek().Type {
	case tokenSpace:
		t.Set(t.Parent())
		t.Next()

		return initParse
	case tokenSlash:
		return parseAlphanum
	default:
		t.Set(t.Parent())

		return nil
	}
}

func parseAlphanumSymbols(t *parse.Tree[token, byte]) parse.ParseFn[token, byte] {
	t.Node(t.Next())

	switch t.Peek().Type {
	case tokenAlphanum:
		t.Node(t.Next())
		t.Set(t.Parent().Parent)

		return parseAlphanum
	default:
		item := t.Next()
		item.Type = tokenError
		t.Node(item)

		return initParse
	}
}

func parseAlphanum(t *parse.Tree[token, byte]) parse.ParseFn[token, byte] {
	switch t.Peek().Type {
	case tokenAlphanum:
		t.Node(t.Next())

		return parseAlphanum
	case tokenComma, tokenDash, tokenSlash:
		return parseAlphanumSymbols
	case tokenSpace:
		t.Set(t.Parent())
		t.Next()

		return initParse
	}

	return nil
}
