package ghttp

import "net/http"

// Endpoints is a wrapper for a set of Handlers
type Endpoints struct {
	E map[string]http.HandlerFunc
}

// NewEndpoints initializes an Endpoints object
func NewEndpoints() *Endpoints {
	return &Endpoints{
		E: make(map[string]http.HandlerFunc),
	}
}

// Set adds the handlers to the Endpoints map
func (e *Endpoints) Set(handlers ...Handler) {
	for _, h := range handlers {
		e.E[h.Path] = h.Fn
	}
}

// Delete removes the handlers from the Endpoints map
func (e *Endpoints) Delete(paths ...string) {
	for _, p := range paths {
		delete(e.E, p)
	}
}
