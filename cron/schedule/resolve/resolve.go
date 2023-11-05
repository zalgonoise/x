package resolve

type Everytime struct{}

func (s Everytime) Resolve(_ int) int {
	return 0
}

type FixedSchedule struct {
	Max int
	At  int
}

func (s FixedSchedule) Resolve(value int) int {
	return diff(value, s.At, s.At, s.Max)
}

type RangeSchedule struct {
	Max  int
	From int
	To   int
}

func (s RangeSchedule) Resolve(value int) int {
	if value > s.From && value < s.To {
		return 0
	}

	return diff(value, s.From, s.To, s.Max)
}

type StepSchedule struct {
	Max   int
	Steps []int
}

func (s StepSchedule) Resolve(value int) int {
	offset := -1

	for i := range s.Steps {
		if offset == -1 {
			offset = diff(value, s.Steps[i], s.Steps[i], s.Max)

			continue
		}

		if n := diff(value, s.Steps[i], s.Steps[i], s.Max); n < offset {
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

func NewStepSchedule(from, to, maximum, frequency int) StepSchedule {
	return StepSchedule{
		Max:   maximum,
		Steps: newValueRange(from, to, frequency),
	}
}

func newValueRange(from, to, frequency int) []int {
	if frequency == 0 || from > to {
		return []int{}
	}

	var r = make([]int, 0, to-from/frequency)

	for i := from; i <= to; i += frequency {
		r = append(r, i)
	}

	return r
}
