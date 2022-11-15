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

func New(w io.Writer) handlers.Handler {
	return &jsonHandler{
		w: w,
	}
}

func (h *jsonHandler) Handle(r records.Record) error {
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
	if r.AttLen() > 0 {
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

func (h *jsonHandler) asMap(attrs []attr.Attr) map[string]interface{} {
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

func (h *jsonHandler) With(attrs ...attr.Attr) handlers.Handler {
	new := &jsonHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  h.levelRef,
		replFn:    h.replFn,
		attrs:     attrs,
	}
	return new
}

func (h *jsonHandler) Enabled(level level.Level) bool {
	if h.levelRef == nil || level.Int() >= h.levelRef.Int() {
		return true
	}
	return false
}

func (h *jsonHandler) WithSource(addSource bool) handlers.Handler {
	new := &jsonHandler{
		w:         h.w,
		addSource: addSource,
		levelRef:  h.levelRef,
		replFn:    h.replFn,
		attrs:     []attr.Attr{},
	}
	copy(new.attrs, h.attrs)
	return new
}

func (h *jsonHandler) WithLevel(level level.Level) handlers.Handler {
	new := &jsonHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  level,
		replFn:    h.replFn,
		attrs:     []attr.Attr{},
	}
	copy(new.attrs, h.attrs)
	return new
}

func (h *jsonHandler) WithReplaceFn(fn func(a attr.Attr) attr.Attr) handlers.Handler {
	new := &jsonHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  h.levelRef,
		replFn:    fn,
		attrs:     []attr.Attr{},
	}
	copy(new.attrs, h.attrs)
	return new
}
