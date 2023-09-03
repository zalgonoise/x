package schedule

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/zalgonoise/parse"
)

func process(t *parse.Tree[token, byte]) (c cronSchedule, err error) {
	if err = validate(t); err != nil {
		return c, err
	}

	nodes := t.List()

	switch len(nodes) {
	case 1:
		return buildException(nodes[0], time.Local), nil
	case 5:
		return cronSchedule{
			Loc:      time.Local,
			min:      buildMinutes(nodes[0]),
			hour:     buildHours(nodes[1]),
			dayMonth: buildMonthDays(nodes[2]),
			month:    buildMonths(nodes[3]),
			dayWeek:  buildWeekdays(nodes[4]),
		}, nil
	default:
		return c, ErrInvalidNumNodes
	}
}

func buildMinutes(node *parse.Node[token, byte]) resolver {
	switch node.Type {
	case tokenStar:
		return processStar(node, 0, 59)
	case tokenAlphanum:
		return processAlphanum(node, 0, 59, nil)
	default:
		return everytime{}
	}
}

func buildHours(node *parse.Node[token, byte]) resolver {
	switch node.Type {
	case tokenStar:
		return processStar(node, 0, 23)
	case tokenAlphanum:
		return processAlphanum(node, 0, 23, nil)
	default:
		return everytime{}
	}
}

func buildMonthDays(node *parse.Node[token, byte]) resolver {
	switch node.Type {
	case tokenStar:
		return processStar(node, 1, 31)
	case tokenAlphanum:
		return processAlphanum(node, 1, 31, nil)
	default:
		return everytime{}
	}
}

func buildMonths(node *parse.Node[token, byte]) resolver {
	switch node.Type {
	case tokenStar:
		return processStar(node, 1, 12)
	case tokenAlphanum:
		return processAlphanum(node, 1, 12, monthsList)
	default:
		return everytime{}
	}
}

func buildWeekdays(node *parse.Node[token, byte]) resolver {
	switch node.Type {
	case tokenStar:
		return processStar(node, 0, 7)
	case tokenAlphanum:
		return processAlphanum(node, 0, 7, weekdaysList)
	default:
		return everytime{}
	}
}

func defaultSchedule(loc *time.Location) cronSchedule {
	return cronSchedule{
		Loc: loc,
		min: fixedSchedule{
			maximum: 59,
			at:      0,
		},
		hour:     everytime{},
		dayMonth: everytime{},
		month:    everytime{},
		dayWeek:  everytime{},
	}
}

func buildException(node *parse.Node[token, byte], loc *time.Location) cronSchedule {
	if node.Type != tokenAt {
		return defaultSchedule(loc)
	}

	if value, ok := getValueFromSymbol(node.Edges[0], 0, 6, exceptionsList); ok {
		switch value {
		// TODO: implement reboot (case 0:)
		case 1:
			return defaultSchedule(loc)
		case 2:
			return cronSchedule{
				Loc: loc,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek:  everytime{},
			}
		case 3:
			return cronSchedule{
				Loc: loc,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: everytime{},
				month:    everytime{},
				dayWeek: fixedSchedule{
					maximum: 6,
					at:      0,
				},
			}
		case 4:
			return cronSchedule{
				Loc: loc,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: fixedSchedule{
					maximum: 31,
					at:      1,
				},
				month:   everytime{},
				dayWeek: everytime{},
			}
		case 5, 6:
			return cronSchedule{
				Loc: loc,
				min: fixedSchedule{
					maximum: 59,
					at:      0,
				},
				hour: fixedSchedule{
					maximum: 23,
					at:      0,
				},
				dayMonth: fixedSchedule{
					maximum: 31,
					at:      1,
				},
				month: fixedSchedule{
					maximum: 12,
					at:      1,
				},
				dayWeek: everytime{},
			}
		}
	}

	return defaultSchedule(loc)
}

func getValueFromSymbol(
	symbol *parse.Node[token, byte], minimum, maximum int, valueList []string,
) (int, bool) {
	if len(symbol.Edges) == 1 {
		value := symbol.Edges[0].Value

		// try to use the value as a number
		if len(value) > 0 && value[0] > '0' && value[0] < '9' {
			num, err := strconv.Atoi(string(value))
			if err == nil {
				return num, true
			}

			return -1, false
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
	}

	return -1, false
}

func processAlphanum(n *parse.Node[token, byte], minimum, maximum int, valueList []string) resolver {
	atValue, err := strconv.Atoi(string(n.Value))
	if err != nil {
		return everytime{}
	}

	if atValue < minimum {
		atValue = minimum
	}

	switch len(n.Edges) {
	case 0:
		return fixedSchedule{
			maximum: maximum,
			at:      atValue,
		}
	default:
		stepValues := make([]int, 0, len(n.Edges)*2)
		every := -1
		rangeEnd := -1

		for i := range n.Edges {
			switch n.Edges[i].Type {
			case tokenComma:
				if value, ok := getValueFromSymbol(n.Edges[i], minimum, maximum, valueList); ok {
					stepValues = append(stepValues, value)
				}

			case tokenDash:
				if value, ok := getValueFromSymbol(n.Edges[i], minimum, maximum, valueList); ok {
					rangeEnd = value
				}

			case tokenSlash:
				if value, ok := getValueFromSymbol(n.Edges[i], minimum, maximum, valueList); ok {
					every = value
				}
			}
		}

		// handle step values only
		if every == -1 && rangeEnd == -1 && len(stepValues) > 0 {
			return stepSchedule{
				maximum: maximum,
				steps:   stepValues,
			}
		}

		// handle range only
		if every == -1 && rangeEnd > 0 && len(stepValues) == 0 {
			return rangeSchedule{
				maximum: maximum,
				from:    atValue,
				to:      rangeEnd,
			}
		}

		// set frequency if unset
		if every < 0 {
			every = 1
		}

		// set end if unset
		if rangeEnd < 0 {
			rangeEnd = maximum
		}

		stepValues = append(stepValues, newValueRange(atValue, rangeEnd, every)...)

		// sort and remove duplicates
		slices.Sort(stepValues)
		slices.Compact(stepValues)

		return stepSchedule{
			maximum: maximum,
			steps:   stepValues,
		}
	}
}

func processStar(n *parse.Node[token, byte], minimum, maximum int) resolver {
	switch len(n.Edges) {
	case 1:
		if n.Edges[0].Type == tokenSlash && len(n.Edges[0].Edges) == 1 {
			stepValue, err := strconv.Atoi(string(n.Edges[0].Edges[0].Value))
			if err != nil {
				return everytime{}
			}

			return newStepSchedule(minimum, maximum, maximum, stepValue)
		}

		fallthrough
	default:
		return everytime{}
	}
}
