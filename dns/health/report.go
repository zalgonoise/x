package health

import "time"

type Report struct {
	*StoreReport `json:"store,omitempty"`
	*DNSReport   `json:"dns,omitempty"`
	*HTTPReport  `json:"http,omitempty"`
	Status       `json:"status,omitempty"`
}

type StoreReport struct {
	Len      int           `json:"num_items,omitempty"`
	Duration time.Duration `json:"query_duration,omitempty"`
	Status   `json:"status,omitempty"`
}

type DNSReport struct {
	Enabled       bool          `json:"is_enabled,omitempty"`
	LocalQuery    time.Duration `json:"local_query_duration,omitempty"`
	ExternalQuery time.Duration `json:"external_query_duration,omitempty"`
	Status        `json:"status,omitempty"`
}

type HTTPReport struct {
	Query  time.Duration `json:"query_duration,omitempty"`
	Status `json:"status,omitempty"`
}

type Status int

const (
	Stopped Status = iota
	Unhealthy
	Running
	Healthy
)

var statusStrings = map[Status]string{
	Stopped:   "stopped",
	Unhealthy: "unhealthy",
	Running:   "running",
	Healthy:   "healthy",
}

func (s Status) String() string {
	return statusStrings[s]
}
