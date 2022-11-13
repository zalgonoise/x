package log

type Handler interface {
	Enabled(level Level) bool
	Handle(Record) error
	With(attrs ...Attr) Handler
}

// Should go in separate packages: handlers/{json,text}/handler.go
//
// type HandlerOptions struct {
// AddSource   bool
// LevelRef    *Level
// ReplaceAttr func(a Attr) Attr
// }
//
// func (ho HandlerOptions) NewJSONHandler(w io.Writer) Handler
// func (ho HandlerOptions) NewTextHandler(w io.Writer) Handler
