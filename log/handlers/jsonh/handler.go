package jsonh

import (
	"errors"
	"io"
	"time"

	json "github.com/goccy/go-json"
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

type jsonHandler struct {
	w         io.Writer
	addSource bool
	levelRef  level.Level
	replFn    func(a attr.Attr) attr.Attr
	attrs     []attr.Attr
}

type jsonRecord struct {
	T     time.Time              `json:"timestamp"`
	M     string                 `json:"message"`
	Level string                 `json:"level"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

// New creates a JSON handler based on the input io.Writer `w`
func New(w io.Writer) handlers.Handler {
	return jsonHandler{
		w: w,
	}
}

// Handle will process the input Record, returning an error if raised
func (h jsonHandler) Handle(r records.Record) error {
	if h.levelRef != nil && r.Level().Int() < h.levelRef.Int() {
		return nil
	}

	var (
		hdata = map[string]interface{}{}
		out   = &jsonRecord{
			T:     r.Time(),
			M:     r.Message(),
			Level: r.Level().String(),
			Data:  map[string]interface{}{},
		}
	)
	if r.AttrLen() > 0 {
		out.Data = h.asMap(r.Attrs())
	}
	if len(h.attrs) > 0 {
		hdata = h.asMap(h.attrs)
	}
	if len(hdata) > 0 {
		for k, v := range hdata {
			out.Data[k] = v
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return err
	}

	n, err := h.w.Write(b)
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrZeroBytes
	}

	return nil
}

func (h jsonHandler) asMap(attrs []attr.Attr) map[string]interface{} {
	var out = map[string]interface{}{}
	for _, a := range attrs {
		if h.replFn != nil {
			a = h.replFn(a)
		}
		if v, ok := (a.Value()).([]attr.Attr); ok {
			out[a.Key()] = h.asMap(v)
			continue
		}
		out[a.Key()] = a.Value()
	}
	return out
}

// With will spawn a copy of this Handler with the input attributes
// `attrs`
func (h jsonHandler) With(attrs ...attr.Attr) handlers.Handler {
	return jsonHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  h.levelRef,
		replFn:    h.replFn,
		attrs:     attrs,
	}
}

// Enabled returns a boolean on whether the Handler is accepting
// records with log level `level`
func (h jsonHandler) Enabled(level level.Level) bool {
	if h.levelRef == nil || level == nil || level.Int() >= h.levelRef.Int() {
		return true
	}
	return false
}

// WithSource will spawn a new copy of this Handler with the setting
// to add a source file+line reference to `addSource` boolean
func (h jsonHandler) WithSource(addSource bool) handlers.Handler {
	return jsonHandler{
		w:         h.w,
		addSource: addSource,
		levelRef:  h.levelRef,
		replFn:    h.replFn,
		attrs:     h.attrs,
	}
}

// WithLevel will spawn a copy of this Handler with the input level `level`
// as a verbosity filter
func (h jsonHandler) WithLevel(level level.Level) handlers.Handler {
	return jsonHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  level,
		replFn:    h.replFn,
		attrs:     h.attrs,
	}
}

// WithReplaceFn will spawn a copy of this Handler with the input attribute
// replace function `fn`
func (h jsonHandler) WithReplaceFn(fn func(a attr.Attr) attr.Attr) handlers.Handler {
	return jsonHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  h.levelRef,
		replFn:    fn,
		attrs:     h.attrs,
	}
}
