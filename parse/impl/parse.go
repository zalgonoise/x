package impl

import (
	"fmt"
	"strings"

	"github.com/zalgonoise/lex"
	"github.com/zalgonoise/x/parse"
)

func TextTemplateParser[C TextToken, T rune](l lex.Lexer[C, T, lex.Item[C, T]]) *parse.Tree[C, T] {
	return parse.New(l, initParse[C, T], (C)(TokenRoot))
}

func initParse[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	for t.Peek().Type != C(TokenEOF) {
		switch t.Peek().Type {
		case (C)(TokenIDENT):
			return parseText[C, T]
		case (C)(TokenLBRACE), (C)(TokenRBRACE):
			return parseTemplate[C, T]
		}
	}
	return nil
}

func parseText[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	if t.Peek().Type == (C)(TokenIDENT) {
		textI := t.Next()
		t.Node(textI.Pos, (C)(TokenText), textI.Value...)
	}
	t.Set(t.Parent())
	return initParse[C, T]
}

func parseTemplate[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	switch t.Peek().Type {
	case (C)(TokenLBRACE):
		tmplI := t.Next()
		t.Node(tmplI.Pos, (C)(TokenTemplate))
		return initParse[C, T]
	case (C)(TokenRBRACE):
		tmplI := t.Next()
		t.Node(tmplI.Pos, (C)(TokenTemplateEnd))
		t.Set(t.Parent().Parent)
	case (C)(TokenEOF):
		t.Node(t.Cur().Pos, (C)(TokenError))
	}
	return initParse[C, T]
}

func processFn[C TextToken, T rune, R string](t *parse.Tree[C, T]) (R, error) {
	var sb = new(strings.Builder)
	var zero R
	for _, n := range t.List() {
		switch n.Type {
		case (C)(TokenText):
			proc, err := processText[C, T, R](n)
			if err != nil {
				return zero, err
			}
			sb.WriteString((string)(proc))
		case (C)(TokenTemplate):
			proc, err := processTemplate[C, T, R](n)
			if err != nil {
				return zero, err
			}
			sb.WriteString((string)(proc))
		}
	}

	return (R)(sb.String()), nil
}

func processText[C TextToken, T rune, R string](n *parse.Node[C, T]) (R, error) {
	var val = make([]rune, len(n.Value), len(n.Value))
	for idx, r := range n.Value {
		val[idx] = (rune)(r)
	}
	return (R)(val), nil
}

func processTemplate[C TextToken, T rune, R string](n *parse.Node[C, T]) (R, error) {
	var sb = new(strings.Builder)
	var ended bool
	var zero R

	sb.WriteString(">>")
	for _, node := range n.Edges {
		switch node.Type {
		case (C)(TokenText):
			proc, err := processText[C, T, R](node)
			if err != nil {
				return zero, err
			}
			sb.WriteString((string)(proc))
		case (C)(TokenTemplate):
			proc, err := processTemplate[C, T, R](node)
			if err != nil {
				return zero, err
			}
			sb.WriteString((string)(proc))
		case (C)(TokenTemplateEnd):
			ended = true
		}
	}
	if !ended {
		return zero, fmt.Errorf("parse error on line: %d", n.Pos)
	}

	sb.WriteString("<<")
	return (R)(sb.String()), nil
}
