package options

var (
	NoType = struct{}{}
)

type intID int

var (
	CfgAdjacencyList Setting = MultiOption(
		GraphList,           // init type
		WithCrossGraphEdges, // override graph matrix settings
	)
	CfgAdjacencyMatrix Setting = MultiOption(
		GraphMatrix,       // init type
		NoCrossGraphEdges, // graph matrix doesn't support cross-graph edges
	)
	CfgLinkedList Setting = MultiOption(
		GraphList,   // init type
		Unweighted,  // adjancy lists do not hold weights
		NonCyclical, // linked lists have no edges
		Directional, // there is no link to the parent
		MaxNodes(1), // linked lists only have one node
	)
	CfgDualLinkedList Setting = MultiOption(
		GraphList,      // init type
		Unweighted,     // adjancy lists do not hold weights
		NonCyclical,    // linked lists have no edges
		NonDirectional, // there is a link back to the parent node
		MaxNodes(1),    // linked lists only have one node
	)
)
