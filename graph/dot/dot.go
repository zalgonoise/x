package dot

import (
	"fmt"
	"strings"

	"github.com/zalgonoise/x/graph/model"
)

type DotGraph[T model.ID, I model.Num] struct {
	cfg *DotConfig
	buf strings.Builder
}

func (g *DotGraph[T, I]) Init() {
	if g.cfg.Direction == "" {
		g.cfg.Direction = string(Directed)
	}
	if g.cfg.WeightKey == "" {
		g.cfg.WeightKey = string(LabelWeight)
	}

	g.buf.WriteString(g.cfg.Direction)
	g.buf.WriteString(` {
`)
}

func (g *DotGraph[T, I]) Add(from, to T, weight I) {
	if weight == 0 {
		return
	}
	g.buf.WriteString(`    `)
	g.buf.WriteString(fmt.Sprint(from))
	g.buf.WriteString(` -> `)
	g.buf.WriteString(fmt.Sprint(to))
	if (weight != 0 && weight != 1) || g.cfg.WeightKey == string(LabelWeight) {
		g.buf.WriteString(` [`)
		g.buf.WriteString(g.cfg.WeightKey)
		g.buf.WriteString(`=`)
		g.buf.WriteString(fmt.Sprint(weight))
		g.buf.WriteString(`]`)
	}
	g.buf.WriteString(`
`)
}

func (g *DotGraph[T, I]) String() string {
	g.buf.WriteString(`}
`)
	return g.buf.String()
}

func New[T model.ID, I model.Num](config ...Setting) *DotGraph[T, I] {
	cfg := &DotConfig{}

	input := MultiOption(config...)

	input.Apply(cfg)

	dot := &DotGraph[T, I]{
		cfg: cfg,
		buf: strings.Builder{},
	}

	dot.Init()

	return dot
}
