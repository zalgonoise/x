package cronlex

import (
	"slices"
	"strconv"
	"strings"

	"github.com/zalgonoise/parse"
	"github.com/zalgonoise/x/cron/schedule/resolve"
)

type Resolver interface {
	Resolve(value int) int
}

type Schedule struct {
	Min      Resolver
	Hour     Resolver
	DayMonth Resolver
	Month    Resolver
	DayWeek  Resolver
}

func Parse(cronString string) (s Schedule, err error) {
	if err = validateCharacters(cronString); err != nil {
		return s, err
	}

	return parse.Run([]byte(cronString), StateFunc, ParseFunc, ProcessFunc)
}

func ProcessFunc(t *parse.Tree[Token, byte]) (s Schedule, err error) {
	if err = Validate(t); err != nil {
		return s, err
	}

	nodes := t.List()

	switch len(nodes) {
	case 1:
		return buildException(nodes[0]), nil
	case 5:
		s = Schedule{
			Min:      buildMinutes(nodes[0]),
			Hour:     buildHours(nodes[1]),
			DayMonth: buildMonthDays(nodes[2]),
			Month:    buildMonths(nodes[3]),
			DayWeek:  buildWeekdays(nodes[4]),
		}

		// convert sundays as 7 into a 0
		if r, ok := s.DayWeek.(resolve.StepSchedule); ok {
			for i := range r.Steps {
				if r.Steps[i] == 7 {
					r.Steps[i] = 0

					slices.Sort(r.Steps)
					s.DayWeek = r
				}
			}
		}

		return s, nil
	default:
		return s, ErrInvalidNumNodes
	}
}

func buildMinutes(node *parse.Node[Token, byte]) Resolver {
	switch node.Type {
	case TokenStar:
		return processStar(node, 0, 59)
	case TokenAlphaNum:
		return processAlphaNum(node, 0, 59, nil)
	default:
		return resolve.Everytime{}
	}
}

func buildHours(node *parse.Node[Token, byte]) Resolver {
	switch node.Type {
	case TokenStar:
		return processStar(node, 0, 23)
	case TokenAlphaNum:
		return processAlphaNum(node, 0, 23, nil)
	default:
		return resolve.Everytime{}
	}
}

func buildMonthDays(node *parse.Node[Token, byte]) Resolver {
	switch node.Type {
	case TokenStar:
		return processStar(node, 1, 31)
	case TokenAlphaNum:
		return processAlphaNum(node, 1, 31, nil)
	default:
		return resolve.Everytime{}
	}
}

func buildMonths(node *parse.Node[Token, byte]) Resolver {
	switch node.Type {
	case TokenStar:
		return processStar(node, 1, 12)
	case TokenAlphaNum:
		return processAlphaNum(node, 1, 12, monthsList)
	default:
		return resolve.Everytime{}
	}
}

func buildWeekdays(node *parse.Node[Token, byte]) Resolver {
	switch node.Type {
	case TokenStar:
		return processStar(node, 0, 7)
	case TokenAlphaNum:
		return processAlphaNum(node, 0, 7, weekdaysList)
	default:
		return resolve.Everytime{}
	}
}

func defaultSchedule() Schedule {
	return Schedule{
		Min: resolve.FixedSchedule{
			Max: 59,
			At:  0,
		},
		Hour:     resolve.Everytime{},
		DayMonth: resolve.Everytime{},
		Month:    resolve.Everytime{},
		DayWeek:  resolve.Everytime{},
	}
}

func buildException(node *parse.Node[Token, byte]) Schedule {
	if node.Type != TokenAt {
		return defaultSchedule()
	}

	if value, ok := getValue(node.Edges[0], 0, 6, exceptionsList); ok {
		switch value {
		// TODO: implement reboot (case 0:)
		case 0: // reboot
			return defaultSchedule()
		case 1: // hourly
			return defaultSchedule()
		case 2: // daily
			return Schedule{
				Min: resolve.FixedSchedule{
					Max: 59,
					At:  0,
				},
				Hour: resolve.FixedSchedule{
					Max: 23,
					At:  0,
				},
				DayMonth: resolve.Everytime{},
				Month:    resolve.Everytime{},
				DayWeek:  resolve.Everytime{},
			}
		case 3: // weekly
			return Schedule{
				Min: resolve.FixedSchedule{
					Max: 59,
					At:  0,
				},
				Hour: resolve.FixedSchedule{
					Max: 23,
					At:  0,
				},
				DayMonth: resolve.Everytime{},
				Month:    resolve.Everytime{},
				DayWeek: resolve.FixedSchedule{
					Max: 6,
					At:  0,
				},
			}
		case 4: // monthly
			return Schedule{
				Min: resolve.FixedSchedule{
					Max: 59,
					At:  0,
				},
				Hour: resolve.FixedSchedule{
					Max: 23,
					At:  0,
				},
				DayMonth: resolve.FixedSchedule{
					Max: 31,
					At:  1,
				},
				Month:   resolve.Everytime{},
				DayWeek: resolve.Everytime{},
			}
		case 5, 6: // yearly, annually
			return Schedule{
				Min: resolve.FixedSchedule{
					Max: 59,
					At:  0,
				},
				Hour: resolve.FixedSchedule{
					Max: 23,
					At:  0,
				},
				DayMonth: resolve.FixedSchedule{
					Max: 31,
					At:  1,
				},
				Month: resolve.FixedSchedule{
					Max: 12,
					At:  1,
				},
				DayWeek: resolve.Everytime{},
			}
		}
	}

	return defaultSchedule()
}

func getValue(
	node *parse.Node[Token, byte], minimum, maximum int, valueList []string,
) (int, bool) {
	value := node.Value

	// try to use the value as a number
	if len(value) > 0 && value[0] >= '0' && value[0] <= '9' {
		num, err := strconv.Atoi(string(value))
		if err == nil {
			return num, true
		}
	}

	// fallback to using it as a string
	v := strings.ToUpper(string(value))
	num := -1

	for idx := range valueList {
		if v == valueList[idx] {
			num = idx
		}
	}

	if num > -1 && num >= minimum && num <= maximum {
		return num, true
	}

	return -1, false
}

func getValueFromSymbol(
	symbol *parse.Node[Token, byte], minimum, maximum int, valueList []string,
) (int, bool) {
	if len(symbol.Edges) == 1 {
		return getValue(symbol.Edges[0], minimum, maximum, valueList)
	}

	return -1, false
}

func processAlphaNum(n *parse.Node[Token, byte], minimum, maximum int, valueList []string) Resolver {
	value, ok := getValue(n, minimum, maximum, valueList)
	if !ok {
		return resolve.Everytime{}
	}

	switch len(n.Edges) {
	case 0:
		return resolve.FixedSchedule{
			Max: maximum,
			At:  value,
		}
	default:
		// there is only one range in the set, do a range-schedule approach
		if len(n.Edges) == 1 && n.Edges[0].Type == TokenDash {
			if to, ok := getValueFromSymbol(n.Edges[0], minimum, maximum, valueList); ok {
				return resolve.RangeSchedule{
					Max:  maximum,
					From: value,
					To:   to,
				}
			}

			return resolve.Everytime{}
		}

		stepValues := make([]int, 0, len(n.Edges)*2)

		// on a mixed scenario we walk through the edges and build a step-schedule out of the combinations provided
		// for reference, TokenDash means a range, TokenSlash means a frequency and TokenComma carries the next value
		//
		// the value variable is reused for this purpose

		for i := range n.Edges {
			switch n.Edges[i].Type {
			case TokenComma:
				// don't leave the initial value dangling when changing Tokens
				if i == 0 {
					stepValues = append(stepValues, value)
				}

				// it's OK to append the (child) value in a comma node
				// even if the next node is a range or a frequency, the same value will be included and repeated values deleted
				//
				// this Token also sets the `cur` variable in case the following Token is a range or frequency
				if v, ok := getValueFromSymbol(n.Edges[i], minimum, maximum, valueList); ok {
					stepValues = append(stepValues, v)

					value = v
				}

			case TokenDash:
				if to, ok := getValueFromSymbol(n.Edges[i], minimum, maximum, valueList); ok {
					stepValues = append(stepValues, buildRange(value, to)...)
				}

			case TokenSlash:
				if freq, ok := getValueFromSymbol(n.Edges[i], minimum, maximum, valueList); ok {
					stepValues = append(stepValues, buildFreq(value, maximum, freq)...)
				}
			}
		}

		slices.Sort(stepValues)
		stepValues = slices.Compact(stepValues)

		return resolve.StepSchedule{
			Max:   maximum,
			Steps: stepValues,
		}
	}
}

func processStar(n *parse.Node[Token, byte], minimum, maximum int) Resolver {
	switch len(n.Edges) {
	case 1:
		if n.Edges[0].Type == TokenSlash && len(n.Edges[0].Edges) == 1 {
			stepValue, err := strconv.Atoi(string(n.Edges[0].Edges[0].Value))
			if err != nil {
				return resolve.Everytime{}
			}

			return resolve.NewStepSchedule(minimum, maximum, maximum, stepValue)
		}

		fallthrough
	default:
		return resolve.Everytime{}
	}
}

func buildRange(from, to int) []int {
	if to < from {
		return []int{}
	}

	out := make([]int, 0, to-from)
	for i := from; i <= to; i++ {
		out = append(out, i)
	}

	return out
}

func buildFreq(base, maximum, freq int) []int {
	if freq == 0 || base > maximum {
		return []int{}
	}

	out := make([]int, 0, maximum-base/freq)
	for i := base; i <= maximum; i += freq {
		out = append(out, i)
	}

	return out
}
