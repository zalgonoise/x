package log

import (
	"sync"
	"time"
)

type Record interface {
	AddAttr(a ...Attr)
	Attr(idx int) Attr
	Attrs() []Attr
	AttLen() int
	Message() string
	Time() time.Time
	Level() Level
}

func NewRecord(t time.Time, level Level, msg string, attrs ...Attr) Record {
	return &record{
		timestamp: t,
		message:   msg,
		level:     level,
		attrs:     attrs,
	}
}

type record struct {
	mu        sync.RWMutex
	timestamp time.Time
	message   string
	level     Level
	attrs     []Attr
}

func (r *record) AddAttr(a ...Attr) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(a) == 0 {
		return
	}
	r.attrs = append(r.attrs, a...)
}

func (r *record) Attr(idx int) Attr {
	r.mu.Lock()
	defer r.mu.Unlock()

	if idx >= len(r.attrs) {
		return nil
	}
	return r.attrs[idx]
}

func (r *record) Attrs() []Attr {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.attrs
}

func (r *record) AttLen() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.attrs)
}

func (r *record) Message() string {
	return r.message
}

func (r *record) Time() time.Time {
	return r.timestamp
}

func (r *record) Level() Level {
	return r.level
}
