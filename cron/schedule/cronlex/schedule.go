package cronlex

type Everytime struct{}

func (s Everytime) Resolve(_ int) int {
	return 0
}

type FixedSchedule struct {
	maximum int
	at      int
}

func (s FixedSchedule) Resolve(value int) int {
	return diff(value, s.at, s.at, s.maximum)
}

type RangeSchedule struct {
	maximum int
	from    int
	to      int
}

func (s RangeSchedule) Resolve(value int) int {
	if value > s.from && value < s.to {
		return 0
	}

	return diff(value, s.from, s.to, s.maximum)
}

type StepSchedule struct {
	maximum int
	steps   []int
}

func (s StepSchedule) Resolve(value int) int {
	offset := -1

	for i := range s.steps {
		if offset == -1 {
			offset = diff(value, s.steps[i], s.steps[i], s.maximum)

			continue
		}

		if n := diff(value, s.steps[i], s.steps[i], s.maximum); n < offset {
			offset = n
		}
	}

	return offset
}

func diff(value, from, to, maximum int) int {
	if value > to {
		return from + maximum - value
	}

	return from - value
}

func newStepSchedule(from, to, maximum, frequency int) StepSchedule {
	return StepSchedule{
		maximum: maximum,
		steps:   newValueRange(from, to, frequency),
	}
}

func newValueRange(from, to, frequency int) []int {
	var r = make([]int, 0, to-from/frequency)

	for i := from; i <= to; i += frequency {
		r = append(r, i)
	}

	return r
}
