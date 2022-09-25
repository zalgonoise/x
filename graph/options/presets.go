package options

var (
	NoType = struct{}{}
)

type intID int

var (
	// CfgAdjacencyList defines the basic configuration of an adjacency list graph
	CfgAdjacencyList Setting = MultiOption(
		GraphList,           // init type
		WithCrossGraphEdges, // override graph matrix settings
	)
	// CfgAdjacencyMatrix defines the basic configuration of an adjacency matrix graph
	CfgAdjacencyMatrix Setting = MultiOption(
		GraphMatrix,       // init type
		NoCrossGraphEdges, // graph matrix doesn't support cross-graph edges
	)
	// CfgKnowledgeGraph defines the basic configuration of a knowledge graph
	CfgKnowledgeGraph Setting = MultiOption(
		GraphKnowledge,      // init type
		WithCrossGraphEdges, // override graph matrix settings
	)
	// CfgLinkedList defines the basic configuration of a linked list type of graph
	CfgLinkedList Setting = MultiOption(
		GraphList,   // init type
		Unweighted,  // adjancy lists do not hold weights
		NonCyclical, // linked lists have no edges
		Directional, // there is no link to the parent
		MaxNodes(1), // linked lists only have one node
	)
	// CfgDualLinkedList defines the basic configuration of a linked list type of graph, that
	// links child nodes to their parents, too
	CfgDualLinkedList Setting = MultiOption(
		GraphList,      // init type
		Unweighted,     // adjancy lists do not hold weights
		NonCyclical,    // linked lists have no edges
		NonDirectional, // there is a link back to the parent node
		MaxNodes(1),    // linked lists only have one node
	)
)
