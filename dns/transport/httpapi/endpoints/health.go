package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/zalgonoise/x/dns/health"
)

type HealthResponse struct {
	Message string
	Report  *health.Report
}

func (e *endpoints) Health(w http.ResponseWriter, r *http.Request) {
	out := e.s.Health()

	w.WriteHeader(200)
	response, _ := json.Marshal(HealthResponse{
		Message: "status and health report",
		Report:  out,
	})
	_, _ = w.Write(response)
}
