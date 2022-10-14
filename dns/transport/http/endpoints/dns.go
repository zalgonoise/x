package endpoints

import (
	"encoding/json"
	"net/http"
)

type DNSResponse struct {
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (e *endpoints) startDNS(w http.ResponseWriter, r *http.Request) {
	err := e.udp.Start()
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(DNSResponse{
			Success: false,
			Message: "failed to start DNS server",
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}
	w.WriteHeader(200)
	response, _ := json.Marshal(DNSResponse{
		Success: true,
		Message: "started DNS server",
	})
	_, _ = w.Write(response)
}
func (e *endpoints) stopDNS(w http.ResponseWriter, r *http.Request) {
	err := e.udp.Stop()
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(DNSResponse{
			Success: false,
			Message: "failed to stop DNS server",
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}
	w.WriteHeader(200)
	response, _ := json.Marshal(DNSResponse{
		Success: true,
		Message: "stopped DNS server",
	})
	_, _ = w.Write(response)
}
func (e *endpoints) reloadDNS(w http.ResponseWriter, r *http.Request) {
	err := e.udp.Stop()
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(DNSResponse{
			Success: false,
			Message: "failed to stop DNS server",
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}
	err = e.udp.Start()
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(DNSResponse{
			Success: false,
			Message: "failed to start DNS server",
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}
	w.WriteHeader(200)
	response, _ := json.Marshal(DNSResponse{
		Success: true,
		Message: "reloaded DNS server",
	})
	_, _ = w.Write(response)
}
