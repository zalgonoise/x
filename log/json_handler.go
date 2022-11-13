package log

import (
	"encoding/json"
	"errors"
	"io"
	"time"
)

var (
	ErrZeroBytes error = errors.New("zero bytes written")
)

type jsonHandler struct {
	w         io.Writer
	addSource bool
	levelRef  Level
	replFn    func(a Attr) Attr
	attrs     []Attr
}

type jsonRecord struct {
	T     time.Time              `json:"timestamp"`
	M     string                 `json:"message"`
	Level string                 `json:"level"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

func NewJSONHandler(w io.Writer) Handler {
	return &jsonHandler{
		w: w,
	}
}

func (h *jsonHandler) Handle(r Record) error {
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

func (h *jsonHandler) asMap(attrs []Attr) map[string]interface{} {
	var out = map[string]interface{}{}
	for _, a := range attrs {
		if h.replFn != nil {
			a = h.replFn(a)
		}
		if v, ok := (a.Value()).([]Attr); ok {
			out[a.Key()] = h.asMap(v)
			continue
		}
		out[a.Key()] = a.Value()
	}
	return out
}

func (h *jsonHandler) With(attrs ...Attr) Handler {
	new := &jsonHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  h.levelRef,
		replFn:    h.replFn,
		attrs:     attrs,
	}
	return new
}

func (h *jsonHandler) Enabled(level Level) bool {
	if h.levelRef == nil || level.Int() >= h.levelRef.Int() {
		return true
	}
	return false
}

func (h *jsonHandler) WithSource(addSource bool) Handler {
	new := &jsonHandler{
		w:         h.w,
		addSource: addSource,
		levelRef:  h.levelRef,
		replFn:    h.replFn,
		attrs:     []Attr{},
	}
	copy(new.attrs, h.attrs)
	return new
}

func (h *jsonHandler) WithLevel(level Level) Handler {
	new := &jsonHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  level,
		replFn:    h.replFn,
		attrs:     []Attr{},
	}
	copy(new.attrs, h.attrs)
	return new
}

func (h *jsonHandler) WithReplaceFn(fn func(a Attr) Attr) Handler {
	new := &jsonHandler{
		w:         h.w,
		addSource: h.addSource,
		levelRef:  h.levelRef,
		replFn:    fn,
		attrs:     []Attr{},
	}
	copy(new.attrs, h.attrs)
	return new
}
