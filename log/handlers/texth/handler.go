package texth

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/handlers"
	"github.com/zalgonoise/x/log/level"
	"github.com/zalgonoise/x/log/records"
)

var (
	// ErrZeroBytes is raised when the `io.Writer` in the handler
	// returns a zero-length of bytes written, when the `Write()`
	// method is called
	ErrZeroBytes error = errors.New("zero bytes written")
)

type textHandler struct {
	w         io.Writer
	addSource bool
	levelRef  level.Level
	replFn    func(a attr.Attr) attr.Attr
	attrs     []attr.Attr
	conf      textHandlerConfig
}

type textHandlerConfig struct {
	wrapperL   rune
	wrapperR   rune
	sepKV      string
	sepAttr    rune
	whiteSpace rune
	timeFmt    string
}

// New creates a text handler based on the input io.Writer `w`
func New(w io.Writer) handlers.Handler {
	return textHandler{
		w: w,
		conf: textHandlerConfig{
			wrapperL:   '[',
			wrapperR:   ']',
			sepKV:      ": ",
			sepAttr:    ';',
			whiteSpace: ' ',
			timeFmt:    time.RFC3339Nano,
		},
	}
}

// Handle will process the input Record, returning an error if raised
func (h textHandler) Handle(r records.Record) error {
	if h.levelRef != nil && r.Level().Int() < h.levelRef.Int() {
		return nil
	}

	var b = &bytes.Buffer{}

	b.WriteRune(h.conf.wrapperL)
	b.WriteString(r.Time().Format(h.conf.timeFmt))
	b.WriteRune(h.conf.wrapperR)
	b.WriteRune(h.conf.wrapperL)
	b.WriteRune(h.conf.whiteSpace)
	b.WriteRune(h.conf.wrapperL)
	b.WriteString(r.Level().String())
	b.WriteRune(h.conf.wrapperR)
	b.WriteRune(h.conf.whiteSpace)
	b.WriteString(r.Message())

	if r.AttrLen() > 0 {
		b.WriteRune(h.conf.whiteSpace)
		b.WriteRune(h.conf.wrapperL)
		b.WriteRune(h.conf.whiteSpace)
		b.WriteString(h.asString(r.Attrs()))
		b.WriteRune(h.conf.whiteSpace)
		b.WriteRune(h.conf.wrapperR)
	}

	n, err := h.w.Write(b.Bytes())
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrZeroBytes
	}

	return nil
}

func (h textHandler) asString(attrs []attr.Attr) string {
	var out = &bytes.Buffer{}
	for idx, a := range attrs {
		if h.replFn != nil {
			a = h.replFn(a)
		}
		if v, ok := (a.Value()).([]attr.Attr); ok {
			out.WriteString(a.Key())
			out.WriteString(h.conf.sepKV)
			out.WriteRune(h.conf.wrapperL)
			out.WriteString(h.asString(v))
			out.WriteRune(h.conf.wrapperR)
			out.WriteRune(h.conf.whiteSpace)
			if idx < len(attrs)-2 {
				out.WriteRune(h.conf.sepAttr)
				out.WriteRune(h.conf.whiteSpace)
			}
			continue
		}
		out.WriteString(a.Key())
		out.WriteString(h.conf.sepKV)
		out.WriteString(fmt.Sprintf("%v", a.Value()))
		out.WriteRune(h.conf.whiteSpace)
		if idx < len(attrs)-2 {
			out.WriteRune(h.conf.sepAttr)
			out.WriteRune(h.conf.whiteSpace)
		}
	}
	return out.String()
}

// With will spawn a copy of this Handler with the input attributes
// `attrs`
func (h textHandler) With(attrs ...attr.Attr) handlers.Handler {
	return textHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  h.levelRef,
		replFn:    h.replFn,
		attrs:     attrs,
	}
}

// Enabled returns a boolean on whether the Handler is accepting
// records with log level `level`
func (h textHandler) Enabled(level level.Level) bool {
	if h.levelRef == nil || level.Int() >= h.levelRef.Int() {
		return true
	}
	return false
}

// WithSource will spawn a new copy of this Handler with the setting
// to add a source file+line reference to `addSource` boolean
func (h textHandler) WithSource(addSource bool) handlers.Handler {
	return textHandler{
		w:         h.w,
		addSource: addSource,
		levelRef:  h.levelRef,
		replFn:    h.replFn,
		attrs:     h.attrs,
	}
}

// WithLevel will spawn a copy of this Handler with the input level `level`
// as a verbosity filter
func (h textHandler) WithLevel(level level.Level) handlers.Handler {
	return textHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  level,
		replFn:    h.replFn,
		attrs:     h.attrs,
	}
}

// WithReplaceFn will spawn a copy of this Handler with the input attribute
// replace function `fn`
func (h textHandler) WithReplaceFn(fn func(a attr.Attr) attr.Attr) handlers.Handler {
	return textHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  h.levelRef,
		replFn:    fn,
		attrs:     h.attrs,
	}
}
