package jsonh

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/zalgonoise/x/log"
)

var (
	ErrZeroBytes error = errors.New("zero bytes written")
)

type jsonHandler struct {
	w         io.Writer
	addSource bool
	levelRef  log.Level
	replFn    func(a log.Attr) log.Attr
	attrs     []log.Attr

	wChan chan log.Record
}

type jsonRecord struct {
	T     time.Time              `json:"timestamp"`
	M     string                 `json:"message"`
	Level string                 `json:"level"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

func New(w io.Writer) log.Handler {
	return &jsonHandler{
		w:     w,
		wChan: make(chan log.Record),
	}
}

func (h *jsonHandler) Handle(r log.Record) error {
	if r.Level().Int() < h.levelRef.Int() {
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

func (h *jsonHandler) asMap(attrs []log.Attr) map[string]interface{} {
	var out = map[string]interface{}{}
	for _, a := range attrs {
		if h.replFn != nil {
			a = h.replFn(a)
		}
		if v, ok := (a.Value()).([]log.Attr); ok {
			out[a.Key()] = h.asMap(v)
			continue
		}
		out[a.Key()] = a.Value()
	}
	return out
}

func (h *jsonHandler) With(attrs ...log.Attr) log.Handler {
	new := *h
	new.attrs = make([]log.Attr, len(h.attrs))
	copy(new.attrs, h.attrs)
	new.attrs = append(new.attrs, attrs...)
	return &new
}

func (h *jsonHandler) Enabled(level log.Level) bool {
	if h.levelRef == nil || level.Int() >= h.levelRef.Int() {
		return true
	}
	return false
}

func (h *jsonHandler) WithSource(addSource bool) log.Handler {
	new := *h
	new.addSource = addSource
	new.attrs = make([]log.Attr, len(h.attrs))
	copy(new.attrs, h.attrs)
	return &new
}

func (h *jsonHandler) WithLevel(level log.Level) log.Handler {
	new := *h
	new.levelRef = level
	new.attrs = make([]log.Attr, len(h.attrs))
	copy(new.attrs, h.attrs)
	return &new
}

func (h *jsonHandler) WithReplaceFn(fn func(a log.Attr) log.Attr) log.Handler {
	new := *h
	new.replFn = fn
	new.attrs = make([]log.Attr, len(h.attrs))
	copy(new.attrs, h.attrs)
	return &new
}
