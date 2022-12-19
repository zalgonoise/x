package codegraph

import (
	"fmt"
	"go/token"

	json "github.com/goccy/go-json"
	cur "github.com/zalgonoise/cur"
)

func New(path string) *WithTokens {
	g := &WithTokens{
		GoFile: &GoFile{},
	}

	if path != "" {
		t, err := Extract(path)
		if err == nil {
			g.Tokens = cur.NewCursor(t)
		}
	}
	return g
}

func (g *GoFile) String() string {
	b, _ := json.Marshal(g)
	return string(b)
}

func NotFound(tok token.Token) error {
	return fmt.Errorf("token %s was required but not found", tok.String())
}
