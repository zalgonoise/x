package level

type Level interface {
	String() string
	Int() int
}

type lv int

const (
	Trace lv = iota
	Debug
	Info
	Warn
	Error
	Fatal
)

func (l lv) String() string {
	return lvValues[l]
}

func (l lv) Int() int {
	return (int)(l)
}

func AsLevel(level string) Level {
	if l, ok := lvKeys[level]; ok {
		return l
	}
	return nil
}
