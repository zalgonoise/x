package handlers

import (
	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/level"
	"github.com/zalgonoise/x/log/records"
)

// Handler describes a logging backend, capable of writing a Record to an
// io.Writer (with its Handle() method).
//
// Beyond this feature, it also exposes methods of copying it with different
// configuration options.
type Handler interface {
	// Enabled returns a boolean on whether the Handler is accepting
	// records with log level `level`
	Enabled(level level.Level) bool
	// Handle will process the input Record, returning an error if raised
	Handle(records.Record) error
	// With will spawn a copy of this Handler with the input attributes
	// `attrs`
	With(attrs ...attr.Attr) Handler

	// WithSource will spawn a new copy of this Handler with the setting
	// to add a source file+line reference to `addSource` boolean
	WithSource(addSource bool) Handler

	// WithLevel will spawn a copy of this Handler with the input level `level`
	// as a verbosity filter
	WithLevel(level level.Level) Handler

	// WithReplaceFn will spawn a copy of this Handler with the input attribute
	// replace function `fn`
	WithReplaceFn(fn func(a attr.Attr) attr.Attr) Handler
}
