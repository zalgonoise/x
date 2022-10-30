package endpoints

import (
	"net/http"

	"github.com/zalgonoise/x/dns/health"
)

type HealthResponse struct {
	Message string         `json:"message,omitempty"`
	Report  *health.Report `json:"report,omitempty"`
}

func (e *endpoints) Health(w http.ResponseWriter, r *http.Request) {
	out := e.s.Health()

	w.WriteHeader(200)
	response, _ := e.enc.Encode(HealthResponse{
		Message: "status and health report",
		Report:  out,
	})
	_, _ = w.Write(response)
}
