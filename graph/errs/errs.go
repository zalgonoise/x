package errs

import "errors"

var (
	AlreadyExists  error = errors.New("already exists")
	NotGraph       error = errors.New("input is not a graph")
	NotNode        error = errors.New("input is not a node")
	DoesNotExist   error = errors.New("target does not exist")
	InvalidType    error = errors.New("invalid input type")
	IDMismatch     error = errors.New("input IDs do not match")
	CyclicalEdge   error = errors.New("cyclical edges not allowed for this type of graph")
	InvalidAdjList error = errors.New("invalid adjancy list config")
)
