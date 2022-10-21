package health

type Report struct {
	*StoreReport `json:"store,omitempty"`
	*DNSReport   `json:"dns,omitempty"`
	*HTTPReport  `json:"http,omitempty"`
	Status       `json:"status,omitempty"`
}

type StoreReport struct {
	Len      int     `json:"num_items,omitempty"`
	Duration float64 `json:"query_ms,omitempty"`
	Status   `json:"status,omitempty"`
}

type DNSReport struct {
	Enabled       bool    `json:"is_enabled,omitempty"`
	LocalQuery    float64 `json:"local_query_ms,omitempty"`
	ExternalQuery float64 `json:"external_query_ms,omitempty"`
	Status        `json:"status,omitempty"`
}

type HTTPReport struct {
	Query  float64 `json:"query_ms,omitempty"`
	Status `json:"status,omitempty"`
}

type Status string

const (
	Stopped   Status = "stopped"
	Unhealthy Status = "unhealthy"
	Running   Status = "running"
	Healthy   Status = "healthy"
)
