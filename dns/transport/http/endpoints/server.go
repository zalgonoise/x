package endpoints

import (
	"context"
	"fmt"
	"net/http"
)

type Server interface {
	Start() error
	Stop() error
}

type server struct {
	ep   HTTPAPI
	port int
	srv  *http.Server
}

func NewServer(api HTTPAPI, port int) Server {
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
	mux.HandleFunc("/dns/start", srv.ep.startDNS)
	mux.HandleFunc("/dns/stop", srv.ep.stopDNS)
	mux.HandleFunc("/dns/reload", srv.ep.reloadDNS)
	mux.HandleFunc("/records/add", srv.ep.addRecord)
	mux.HandleFunc("/records", srv.ep.listRecords)
	mux.HandleFunc("/records/getAddress", srv.ep.getRecordByDomain)
	mux.HandleFunc("/records/getDomains", srv.ep.getRecordByAddress)
	mux.HandleFunc("/records/update", srv.ep.updateRecord)
	mux.HandleFunc("/records/delete", srv.ep.deleteRecord)

	return srv
}

func (s *server) Start() error {
	return http.ListenAndServe(
		fmt.Sprintf(":%v", s.port),
		nil,
	)
}

func (s *server) Stop() error {
	return s.srv.Shutdown(context.Background())
}
