package schedule

type everytime struct{}

func (s everytime) Resolve(_ int) int {
	return 0
}

type fixedSchedule struct {
	maximum int
	at      int
}

func (s fixedSchedule) Resolve(value int) int {
	return diff(value, s.at, s.at, s.maximum)
}

type rangeSchedule struct {
	maximum int
	from    int
	to      int
}

func (s rangeSchedule) Resolve(value int) int {
	if value > s.from && value < s.to {
		return 0
	}

	return diff(value, s.from, s.to, s.maximum)
}

type stepSchedule struct {
	maximum int
	steps   []int
}

func (s stepSchedule) Resolve(value int) int {
	var offset int = -1

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

func newStepSchedule(from, to, maximum, frequency int) stepSchedule {
	return stepSchedule{
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
