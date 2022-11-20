package records

import (
	"time"

	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/level"
)

// Record interface describes the behavior that a Record should have
//
// It expose getter methods for its elements, as well as two helper methods:
//   - `AddAttr()` will return a copy of this Record with the input Attr appended
//     to the existing ones
//   - `AttrLen()` will return the length of the attributes in the record
type Record interface {
	// AddAttr returns a copy of this Record with the input Attr appended to the
	// existing ones
	AddAttr(a ...attr.Attr) Record
	// Attrs returns the slice of Attr associated to this Record
	Attrs() []attr.Attr
	// AttrLen returns the length of the slice of Attr in the Record
	AttrLen() int
	// Message returns the string Message associated to this Record
	Message() string
	// Time returns the time.Time timestamp associated to this Record
	Time() time.Time
	// Level returns the level.Level level associated to this Record
	Level() level.Level
}

// New will return a Record based on the input time.Time `t`, level.Level `lv`,
// message string `msg` and attributes `attrs`.
//
// If the input time is zero or Unix Time zero, it will set the Records time as now.
// If the level is nil, level.Info is set instead
// If there are nil attr.Attr values, they are dismissed
//
// A Record is based on an immutable data structure
func New(t time.Time, lv level.Level, msg string, attrs ...attr.Attr) Record {
	if t.IsZero() || t == time.Unix(0, 0) {
		t = time.Now()
	}
	if lv == nil {
		lv = level.Info
	}

	as := []attr.Attr{}
	for _, a := range attrs {
		if a != nil {
			as = append(as, a)
		}
	}
	return record{
		timestamp: t,
		message:   msg,
		level:     lv,
		attrs:     as,
	}
}

type record struct {
	timestamp time.Time
	message   string
	level     level.Level
	attrs     []attr.Attr
}

// AddAttr returns a copy of this Record with the input Attr appended to the
// existing ones
func (r record) AddAttr(attrs ...attr.Attr) Record {
	as := r.attrs
	for _, a := range attrs {
		if a != nil {
			as = append(as, a)
		}
	}
	return record{
		timestamp: r.timestamp,
		message:   r.message,
		level:     r.level,
		attrs:     as,
	}
}

// Attrs returns the slice of Attr associated to this Record
func (r record) Attrs() []attr.Attr {
	return r.attrs
}

// AttrLen returns the length of the slice of Attr in the Record
func (r record) AttrLen() int {
	return len(r.attrs)
}

// Message returns the string Message associated to this Record
func (r record) Message() string {
	return r.message
}

// Time returns the time.Time timestamp associated to this Record
func (r record) Time() time.Time {
	return r.timestamp
}

// Level returns the level.Level level associated to this Record
func (r record) Level() level.Level {
	return r.level
}
