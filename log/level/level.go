package level

// Level interface describes the behavior that a log level should have
//
// It must provide methods to be casted as a string or as an int
type Level interface {
	// String returns the level as a string
	String() string
	// Int returns the level as an int
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
	// Trace represents log level 0
	Trace Level = lTrace
	// Debug represents log level 1
	Debug Level = lDebug
	// Info represents log level 2
	Info Level = lInfo
	// Warn represents log level 3
	Warn Level = lWarn
	// Error represents log level 4
	Error Level = lError
	// Fatal represents log level 5
	Fatal Level = lFatal
)

// String returns the level as a string
func (l lv) String() string {
	return lvValues[l]
}

// Int returns the level as an int
func (l lv) Int() int {
	return (int)(l)
}

// AsLevel converts an input string to a Level, returning nil if
// invalid
func AsLevel(level string) Level {
	if l, ok := lvKeys[level]; ok {
		return l
	}
	return nil
}
