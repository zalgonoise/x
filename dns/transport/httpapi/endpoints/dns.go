package endpoints

import (
	"net/http"

	"github.com/zalgonoise/x/dns/transport/httpapi"
)

func (e *endpoints) StartDNS(w http.ResponseWriter, r *http.Request) {
	var err error

	go func() {
		err = e.UDP.Start()
	}()

	if err != nil {
		w.WriteHeader(500)
		response, _ := e.enc.Encode(httpapi.DNSResponse{
			Success: false,
			Message: "failed to start DNS server",
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}
	w.WriteHeader(200)
	response, _ := e.enc.Encode(httpapi.DNSResponse{
		Success: true,
		Message: "started DNS server",
	})
	_, _ = w.Write(response)
}

func (e *endpoints) StopDNS(w http.ResponseWriter, r *http.Request) {
	err := e.UDP.Stop()
	if err != nil {
		w.WriteHeader(500)
		response, _ := e.enc.Encode(httpapi.DNSResponse{
			Success: false,
			Message: "failed to stop DNS server",
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}
	w.WriteHeader(200)
	response, _ := e.enc.Encode(httpapi.DNSResponse{
		Success: true,
		Message: "stopped DNS server",
	})
	_, _ = w.Write(response)
}

func (e *endpoints) ReloadDNS(w http.ResponseWriter, r *http.Request) {
	err := e.UDP.Stop()
	if err != nil {
		w.WriteHeader(500)
		response, _ := e.enc.Encode(httpapi.DNSResponse{
			Success: false,
			Message: "failed to stop DNS server",
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	go func() {
		err = e.UDP.Start()
	}()

	if err != nil {
		w.WriteHeader(500)
		response, _ := e.enc.Encode(httpapi.DNSResponse{
			Success: false,
			Message: "failed to start DNS server",
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}
	w.WriteHeader(200)
	response, _ := e.enc.Encode(httpapi.DNSResponse{
		Success: true,
		Message: "reloaded DNS server",
	})
	_, _ = w.Write(response)
}
