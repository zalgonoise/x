package schedule

type fixedSchedule struct {
	maximum int8
	at      int8
}

func (s fixedSchedule) Resolve(value int8) int8 {
	return diff(value, s.at, s.at, s.maximum)
}

type rangeSchedule struct {
	maximum int8
	from    int8
	to      int8
}

func (s rangeSchedule) Resolve(value int8) int8 {
	if value > s.from && value < s.to {
		return 0
	}

	return diff(value, s.from, s.to, s.maximum)
}

type stepSchedule struct {
	maximum int8
	steps   []int8
}

func (s stepSchedule) Resolve(value int8) int8 {
	var offset int8 = -1

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

func diff(value, from, to, maximum int8) int8 {
	if value > to {
		return from + maximum - value
	}

	return from - value
}

func newStepSchedule(from, to, maximum, frequency int8) stepSchedule {
	var r = make([]int8, 0, to-from/frequency)

	for i := from; i < maximum; i += frequency {
		r = append(r, i)
	}

	return stepSchedule{
		maximum: maximum,
		steps:   r,
	}
}
