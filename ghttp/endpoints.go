package ghttp

// Endpoints maps URL paths to a list of Handlers
type Endpoints map[string][]Handler

// NewEndpoints initializes an Endpoints object
func NewEndpoints() Endpoints {
	return make(map[string][]Handler)
}

// Set adds the handlers to the Endpoints map
func (e Endpoints) Set(handlers ...Handler) {
	for _, h := range handlers {
		e[h.Path] = append(e[h.Path], h)
	}
}

// Delete removes the handlers from the Endpoints map
func (e Endpoints) Delete(paths ...string) {
	for _, p := range paths {
		delete(e, p)
	}
}
