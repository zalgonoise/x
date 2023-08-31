package schedule

import (
	"testing"

	"github.com/zalgonoise/parse"
)

func TestParser(t *testing.T) {
	t.Log(parse.Run([]byte("5,3 * * * *"), initState, initParse, process))
}
