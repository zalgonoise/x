package ghttp

// Endpoints is a wrapper for a set of Handlers
type Endpoints struct {
	E map[string][]Handler
}

// NewEndpoints initializes an Endpoints object
func NewEndpoints() Endpoints {
	return Endpoints{
		E: make(map[string][]Handler),
	}
}

// Set adds the handlers to the Endpoints map
func (e Endpoints) Set(handlers ...Handler) {
	for _, h := range handlers {
		e.E[h.Path] = append(e.E[h.Path], h)
	}
}

// Delete removes the handlers from the Endpoints map
func (e Endpoints) Delete(paths ...string) {
	for _, p := range paths {
		delete(e.E, p)
	}
}
