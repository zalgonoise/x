package protofile

import (
	"github.com/zalgonoise/gio"
	"github.com/zalgonoise/lex"
	"github.com/zalgonoise/parse"
)

func Run[C ProtoToken, T byte, R string](r gio.Reader[T]) (R, error) {
	var rootEOF C
	l := (lex.Emitter[C, T])(lex.NewBuffer(initState[C, T], r))
	t := parse.New(l, initParse[C, T], rootEOF)
	t.Parse()
	return processFn[C, T, R](t)
}
