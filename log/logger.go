package log

import (
	"context"
	"os"
	"time"
)

type Logger interface {
	Trace(msg string, attrs ...Attr)
	Debug(msg string, attrs ...Attr)
	Info(msg string, attrs ...Attr)
	Warn(msg string, attrs ...Attr)
	Error(msg string, attrs ...Attr)
	Fatal(msg string, attrs ...Attr)
	Log(level Level, msg string, attrs ...Attr)
	Enabled(level Level) bool
	Handler() Handler
	With(attrs ...Attr) Logger
}

// TODO: refactor into packages ; avoiding cycles with a factory (?)

var stdLogger = New(NewJSONHandler(os.Stderr))

type logger struct {
	h        Handler
	attrs    []Attr
	levelRef Level
}

func New(h Handler) Logger {
	return &logger{
		h: h,
	}
}

func Default() Logger {
	return stdLogger
}

func With(attrs ...Attr) Logger {
	return &logger{
		h:        stdLogger.Handler(),
		levelRef: (stdLogger).(*logger).levelRef,
		attrs:    attrs,
	}
}

func (l *logger) WithLevel(level Level) Logger {
	new := &logger{
		h:        l.h,
		levelRef: level,
	}
	copy(new.attrs, l.attrs)
	return new
}

func (l *logger) With(attrs ...Attr) Logger {
	return &logger{
		h:        l.h,
		levelRef: l.levelRef,
		attrs:    attrs,
	}
}

func (l *logger) Log(level Level, msg string, attrs ...Attr) {
	if msg == "" {
		return
	}
	if level == nil {
		level = LInfo
	}

	rAttr := append(attrs, l.attrs...)
	r := NewRecord(time.Now(), level, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Trace(msg string, attrs ...Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := NewRecord(time.Now(), LTrace, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Debug(msg string, attrs ...Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := NewRecord(time.Now(), LDebug, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Info(msg string, attrs ...Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := NewRecord(time.Now(), LInfo, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Warn(msg string, attrs ...Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := NewRecord(time.Now(), LWarn, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Error(msg string, attrs ...Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := NewRecord(time.Now(), LError, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Fatal(msg string, attrs ...Attr) {
	if msg == "" {
		return
	}

	rAttr := append(attrs, l.attrs...)
	r := NewRecord(time.Now(), LFatal, msg, rAttr...)
	_ = l.h.Handle(r)
}
func (l *logger) Enabled(level Level) bool {
	if l.levelRef == nil || level.Int() >= l.levelRef.Int() {
		return true
	}
	return false
}
func (l *logger) Handler() Handler {
	return l.h
}

func InContext(ctx context.Context, logger Logger) context.Context {
	return nil
}
func From(ctx context.Context) Logger {
	return nil
}
