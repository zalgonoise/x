package protofile

import (
	"github.com/zalgonoise/gio"
	"github.com/zalgonoise/parse"
)

func Parse[C ProtoToken, T byte, R int](r gio.Reader[T]) (R, error) {
	return parse.Parse(r, initState[C, T], initParse[C, T], processFn[C, T, R])
}
