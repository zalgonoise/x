package dot

type Direction string

const (
	Directed   Direction = "digraph"
	Undirected Direction = "graph"
)

type WeightString string

const (
	LabelWeight    WeightString = "label"
	DistanceWeight WeightString = "weight"
)

type Setting interface {
	Apply(*DotConfig)
}

type DotConfig struct {
	Direction string
	WeightKey string
}

func (d Direction) Apply(cfg *DotConfig) {
	cfg.Direction = string(d)
}

func (d WeightString) Apply(cfg *DotConfig) {
	cfg.Direction = string(d)
}
