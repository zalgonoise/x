package options

type (
	// TypeSetting defines the graph type (adjacency list, adjacency matrix and knowledge graph)
	TypeSetting int
	// DirectionSetting defines whether the graph is directed or undirected
	DirectionSetting int
	// CycleSetting defines whether the graph allows cycles or not
	CycleSetting int
	// WeightedEdges defines whether the graph allows edge weights or not
	WeightedEdges int
	// Mutability defines whether the graph can be altered (update / delete actions) or not
	Mutability int
	// WritePrivilege defines whether the graph can be written to (create / update / delete)
	// once connected to a parent, or not
	WritePrivilege int
	// NodeLimit defines the number of nodes the graph can hold
	NodeLimit int
	// DepthLimit defines the maximum depth a graph can have (depth 3: root->a->b->c)
	DepthLimit int
	// WeightAsDistance defines whether the dot output sets the weight as a
	// weight property (as distance) or a label property
	WeightAsDistance int
	// CrossGraphEdges defines whether the graph allows edges across graphs (edges connecting nodes
	// that do not share the same parent graph)
	CrossGraphEdges int
)

const (
	GraphList TypeSetting = iota
	GraphMatrix
	GraphKnowledge
)
const (
	Directional DirectionSetting = iota
	NonDirectional
)
const (
	Cyclical CycleSetting = iota
	NonCyclical
)
const (
	Weighted WeightedEdges = iota
	Unweighted
)
const (
	Mutable Mutability = iota
	Immutable
)
const (
	ReadWrite WritePrivilege = iota
	ReadOnly
)

const (
	LabelWeight WeightAsDistance = iota
	DistanceWeight
)

const (
	WithCrossGraphEdges CrossGraphEdges = iota
	NoCrossGraphEdges
)

func MaxNodes(v int) Setting {
	if v < 0 {
		v = 0
	}

	s := NodeLimit(v)
	return &s
}

func MaxDepth(v int) Setting {
	if v < 0 {
		v = 0
	}

	s := DepthLimit(v)
	return &s
}

func (s TypeSetting) Apply(c *GraphConfig) {
	c.GraphType = s
	if s == GraphMatrix {
		NoCrossGraphEdges.Apply(c)
		return
	}
	WithCrossGraphEdges.Apply(c)
}

func (s DirectionSetting) Apply(c *GraphConfig) {
	if s == NonDirectional {
		c.IsNonDirectional = true
		return
	}
	c.IsNonDirectional = false
}

func (s CycleSetting) Apply(c *GraphConfig) {
	if s == NonCyclical {
		c.IsNonCyclical = true
		return
	}
	c.IsNonCyclical = false
}

func (s WeightedEdges) Apply(c *GraphConfig) {
	if s == Unweighted {
		c.IsUnweighted = true
		return
	}
	c.IsUnweighted = false
}

func (s Mutability) Apply(c *GraphConfig) {
	if s == Immutable {
		c.Immutable = true
		return
	}
	c.Immutable = false
}

func (s WritePrivilege) Apply(c *GraphConfig) {
	if s == ReadOnly {
		c.ReadOnly = true
		return
	}
	c.ReadOnly = false
}

func (s NodeLimit) Apply(c *GraphConfig) {
	if s <= 0 {
		c.MaxNodes = 0
		return
	}
	c.MaxNodes = int(s)
}

func (s DepthLimit) Apply(c *GraphConfig) {
	if s <= 0 {
		c.MaxDepth = 0
		return
	}
	c.MaxDepth = int(s)
}

func (s WeightAsDistance) Apply(c *GraphConfig) {
	if s == DistanceWeight {
		c.WeightAsDistance = true
		return
	}
	c.WeightAsDistance = false
}

func (s CrossGraphEdges) Apply(c *GraphConfig) {
	if s == NoCrossGraphEdges {
		c.NoCrossGraphEdges = true
		return
	}
	c.NoCrossGraphEdges = false
}
