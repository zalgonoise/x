package log

import (
	"time"

	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/level"
	"github.com/zalgonoise/x/log/records"
)

type Printer interface {
	Trace(msg string, attrs ...attr.Attr)
	Debug(msg string, attrs ...attr.Attr)
	Info(msg string, attrs ...attr.Attr)
	Warn(msg string, attrs ...attr.Attr)
	Error(msg string, attrs ...attr.Attr)
	Fatal(msg string, attrs ...attr.Attr)
	Log(level level.Level, msg string, attrs ...attr.Attr)
}

func (l *logger) Log(lv level.Level, msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}
	if lv == nil {
		lv = level.Info
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), lv, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Trace(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Trace, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Debug(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Debug, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Info(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Info, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Warn(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Warn, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Error(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Error, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Fatal(msg string, attrs ...attr.Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := records.New(time.Now(), level.Fatal, msg, rAttr...)
	_ = l.h.Handle(r)
}
