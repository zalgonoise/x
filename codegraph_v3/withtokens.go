package codegraph

import (
	cur "github.com/zalgonoise/cur"
)

type WithTokens struct {
	*GoFile
	Tokens cur.Cursor[GoToken]
}
