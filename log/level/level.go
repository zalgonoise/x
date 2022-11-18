package level

type Level interface {
	String() string
	Int() int
}

type lv int

const (
	lTrace lv = iota
	lDebug
	lInfo
	lWarn
	lError
	lFatal
)

var (
	Trace Level = lTrace
	Debug Level = lDebug
	Info  Level = lInfo
	Warn  Level = lWarn
	Error Level = lError
	Fatal Level = lFatal
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
