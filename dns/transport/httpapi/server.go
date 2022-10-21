package httpapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zalgonoise/x/dns/transport/httpapi/endpoints"
)

type Server interface {
	Start() error
	Stop() error
}

type server struct {
	ep   endpoints.HTTPAPI
	port int
	srv  *http.Server
}

func NewServer(api endpoints.HTTPAPI, port int) Server {
	mux := http.NewServeMux()
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: mux,
	}
	srv := &server{
		ep:   api,
		port: port,
		srv:  httpSrv,
	}
	mux.HandleFunc("/dns/start", srv.ep.StartDNS)
	mux.HandleFunc("/dns/stop", srv.ep.StopDNS)
	mux.HandleFunc("/dns/reload", srv.ep.ReloadDNS)
	mux.HandleFunc("/records/add", srv.ep.AddRecord)
	mux.HandleFunc("/records", srv.ep.ListRecords)
	mux.HandleFunc("/records/getAddress", srv.ep.GetRecordByDomain)
	mux.HandleFunc("/records/getDomains", srv.ep.GetRecordByAddress)
	mux.HandleFunc("/records/update", srv.ep.UpdateRecord)
	mux.HandleFunc("/records/delete", srv.ep.DeleteRecord)
	mux.HandleFunc("/health", srv.ep.Health)

	return srv
}

func (s *server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *server) Stop() error {
	var (
		rw  = &responseWriter{}
		err error
	)
	s.ep.StopDNS(rw, &http.Request{})

	if rw.header != 200 {
		err = fmt.Errorf("%s", rw.response)
	}

	httpErr := s.srv.Shutdown(context.Background())
	if err == nil && httpErr != nil {
		err = httpErr
		httpErr = nil
	}
	if httpErr != nil && err != nil {
		err = fmt.Errorf("http: %v ; udp: %w", httpErr, err)
	}

	return err
}
