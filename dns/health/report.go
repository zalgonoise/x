package health

// Report contains a full perspective of the app's health
//
// Consists of a StoreReport, a DNSReport and a HTTPReport,
// as well as a general Status for the app
type Report struct {
	*StoreReport `json:"store,omitempty"`
	*DNSReport   `json:"dns,omitempty"`
	*HTTPReport  `json:"http,omitempty"`
	Status       `json:"status,omitempty"`
}

// StoreReport defines the health of the embeded store in this service
// by returning information on the number of items, the duration it took to
// perform a store.List operation and its derived status
type StoreReport struct {
	Len      int     `json:"num_items,omitempty"`
	Duration float64 `json:"query_ms,omitempty"`
	Status   `json:"status,omitempty"`
}

// DNSReport defines the health of the embeded DNS in this service
// by returning information on whether it is enabled, the duration of
// a local DNS query in milliseconds, the duration of an external
// DNS query in milliseconds and its derived status
type DNSReport struct {
	Enabled       bool    `json:"is_enabled,omitempty"`
	LocalQuery    float64 `json:"local_query_ms,omitempty"`
	ExternalQuery float64 `json:"external_query_ms,omitempty"`
	Status        `json:"status,omitempty"`
}

// HTTPReport defines the health of the embeded HTTP API in this service
// by returning information on the duration of an API request in
// milliseconds and its derived status
type HTTPReport struct {
	Query  float64 `json:"query_ms,omitempty"`
	Status `json:"status,omitempty"`
}

// Status is a reserved type to list status types
//
// This is not an enum as it is used so little,
// the cost of making it an int with converters was looking like
// too much clutter. It's only actually needed once and it does not
// seem that much of a long condition
type Status string

const (
	// Stopped shows that the service is not running
	Stopped Status = "stopped"
	// Unhealthy shows that the service is running, but presenting issues
	Unhealthy Status = "unhealthy"
	// Running shows that the service is generally running
	Running Status = "running"
	// Healthy shows that the service is running and responding to requests correctly
	Healthy Status = "healthy"
)
