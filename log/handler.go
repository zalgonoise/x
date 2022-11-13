package log

type Handler interface {
	Enabled(level Level) bool
	Handle(Record) error
	With(attrs ...Attr) Handler
}
