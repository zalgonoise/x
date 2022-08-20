package errs

import "errors"

var (
	AlreadyExists   error = errors.New("already exists")
	NotGraph        error = errors.New("input is not a graph")
	NotNode         error = errors.New("input is not a node")
	DoesNotExist    error = errors.New("target does not exist")
	InvalidType     error = errors.New("invalid input type")
	IDMismatch      error = errors.New("input IDs do not match")
	CyclicalEdge    error = errors.New("cyclical edges not allowed for this type of graph")
	InvalidAdjList  error = errors.New("invalid adjancy list config")
	Immutable       error = errors.New("cannot update or delete an immutable graph")
	ReadOnly        error = errors.New("cannot write / add nodes to a read-only graph")
	MaxNodesReached error = errors.New("reached maximum amount of nodes in this graph")
	MaxDepthReached error = errors.New("reached maximum depth for this graph")
)
