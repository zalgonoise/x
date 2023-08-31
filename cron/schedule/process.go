package schedule

import (
	"strings"

	"github.com/zalgonoise/parse"
)

func process(t *parse.Tree[token, byte]) (string, error) {
	sb := new(strings.Builder)
	nodes := t.List()

	for i := range nodes {
		switch nodes[i].Type {
		case tokenStar:
			sb.WriteString("everything: ")
			sb.Write(nodes[i].Value)
			sb.WriteString(" ;")
		case tokenAlphanum:
			processAlphanum(sb, nodes[i])
		case tokenComma:
			sb.WriteString("and: ")
			sb.Write(nodes[i].Value)
			sb.WriteString(" ;")
		case tokenDash:
			sb.WriteString("to: ")
			sb.Write(nodes[i].Value)
			sb.WriteString(" ;")
		case tokenSlash:
			sb.WriteString("by: ")
			sb.Write(nodes[i].Value)
			sb.WriteString(" ;")
		case tokenAt:
			sb.WriteString("exception: ")
			sb.Write(nodes[i].Value)
			sb.WriteString(" ;")
		default:
			break
		}
	}

	return sb.String(), nil
}

func processAlphanum(sb *strings.Builder, n *parse.Node[token, byte]) {
	sb.WriteString("alphanum: ")
	sb.Write(n.Value)

	for i := 0; i < len(n.Edges); i++ {
		switch n.Edges[i].Type {
		case tokenComma:
			sb.WriteString(" and ")
		case tokenDash:
			sb.WriteString(" to ")
		case tokenSlash:
			sb.WriteString(" by ")
		}

		if len(n.Edges[i].Edges) > 0 {
			sb.Write(n.Edges[i].Edges[0].Value)
		}
	}

	sb.WriteString(" ; ")
}
