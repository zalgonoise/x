package records

import (
	"time"

	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/level"
)

type Record interface {
	AddAttr(a ...attr.Attr) Record
	Attr(idx int) attr.Attr
	Attrs() []attr.Attr
	AttLen() int
	Message() string
	Time() time.Time
	Level() level.Level
}

func New(t time.Time, lv level.Level, msg string, attrs ...attr.Attr) Record {
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

func (r record) Attr(idx int) attr.Attr {
	if idx >= len(r.attrs) {
		return nil
	}
	return r.attrs[idx]
}

func (r record) Attrs() []attr.Attr {
	return r.attrs
}

func (r record) AttLen() int {
	return len(r.attrs)
}

func (r record) Message() string {
	return r.message
}

func (r record) Time() time.Time {
	return r.timestamp
}

func (r record) Level() level.Level {
	return r.level
}
