package impl

import (
	"fmt"
	"strings"

	"github.com/zalgonoise/x/parse"
)

func processFn[C TextToken, T rune, R string](t *parse.Tree[C, T]) (R, error) {
	var sb = new(strings.Builder)
	for _, n := range t.List() {
		switch n.Type {
		case (C)(TokenIDENT):
			proc, err := processText[C, T, R](n)
			if err != nil {
				return (R)(sb.String()), err
			}
			sb.WriteString((string)(proc))
		case (C)(TokenLBRACE):
			proc, err := processTemplate[C, T, R](n)
			if err != nil {
				return (R)(sb.String()), err
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

	sb.WriteString(">>")
	for _, node := range n.Edges {
		switch node.Type {
		case (C)(TokenIDENT):
			proc, err := processText[C, T, R](node)
			if err != nil {
				return (R)(sb.String()), err
			}
			sb.WriteString((string)(proc))
		case (C)(TokenLBRACE):
			proc, err := processTemplate[C, T, R](node)
			if err != nil {
				return (R)(sb.String()), err
			}
			sb.WriteString((string)(proc))
		case (C)(TokenRBRACE):
			ended = true
		}
	}
	if !ended {
		return (R)(sb.String()), fmt.Errorf("parse error on line: %d", n.Pos)
	}

	sb.WriteString("<<")
	return (R)(sb.String()), nil
}
