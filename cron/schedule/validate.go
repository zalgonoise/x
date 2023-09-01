package schedule

import (
	"errors"
	"fmt"

	"github.com/zalgonoise/parse"
	"github.com/zalgonoise/x/errs"
)

const (
	numNodesWithOverride = 1
	defaultNumNodes      = 5

	errDomain = errs.Domain("x/cron")

	ErrInvalid   = errs.Kind("invalid")
	ErrErrorType = errs.Kind("error-type")

	ErrNumNodes  = errs.Entity("number of nodes")
	ErrNode      = errs.Entity("node")
	ErrNodeType  = errs.Entity("node type")
	ErrNumEdges  = errs.Entity("number of edges")
	ErrFrequency = errs.Entity("frequency")
)

var (
	ErrInvalidNumNodes  = errs.New(errDomain, ErrInvalid, ErrNumNodes)
	ErrErrorTypeNode    = errs.New(errDomain, ErrErrorType, ErrNode)
	ErrInvalidNodeType  = errs.New(errDomain, ErrInvalid, ErrNodeType)
	ErrInvalidNumEdges  = errs.New(errDomain, ErrInvalid, ErrNumEdges)
	ErrInvalidFrequency = errs.New(errDomain, ErrInvalid, ErrFrequency)
)

func validate(t *parse.Tree[token, byte]) error {
	nodes := t.List()

	switch len(nodes) {
	case 1:
		return validateOverride(nodes[0])
	case 5:
		return errors.Join(
			validateMinutes(nodes[0]),
			validateHours(nodes[1]),
			validateMonthDays(nodes[2]),
			validateMonths(nodes[3]),
			validateWeekDays(nodes[4]),
		)
	default:
		return fmt.Errorf("%w: %d", ErrInvalidNumNodes, len(nodes))
	}
}

func validateOverride(node *parse.Node[token, byte]) error {
	if node.Type != tokenAt {
		return fmt.Errorf("%w: %T -- %v", ErrInvalidNodeType, node.Type, node.Value)
	}

	if len(node.Edges) != 1 {
		return fmt.Errorf("%w: %d", ErrInvalidNumEdges, len(node.Edges))
	}

	frequency := string(node.Edges[0].Value)

	switch frequency {
	case "yearly", "annually", "monthly", "weekly", "daily", "hourly", "reboot":
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidFrequency, frequency)
	}
}

// TODO: implement validateMinutes
func validateMinutes(node *parse.Node[token, byte]) error {
	return nil
}

// TODO: implement validateHours
func validateHours(node *parse.Node[token, byte]) error {
	return nil
}

// TODO: implement validateMonthDays
func validateMonthDays(node *parse.Node[token, byte]) error {
	return nil
}

// TODO: implement validateMonths
func validateMonths(node *parse.Node[token, byte]) error {
	return nil
}

// TODO: implement validateWeekDays
func validateWeekDays(node *parse.Node[token, byte]) error {
	return nil
}
