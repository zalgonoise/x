package endpoints

import (
	"fmt"
	"net/http"
)

type Server interface {
	Start() error
}

type server struct {
	ep   HTTPAPI
	port int
}

func NewServer(api HTTPAPI, port int) Server {
	srv := &server{
		ep:   api,
		port: port,
	}
	http.HandleFunc("/dns/start", srv.ep.startDNS)
	http.HandleFunc("/dns/stop", srv.ep.stopDNS)
	http.HandleFunc("/dns/reload", srv.ep.reloadDNS)
	http.HandleFunc("/records/add", srv.ep.addRecord)
	http.HandleFunc("/records", srv.ep.listRecords)
	http.HandleFunc("/records/getAddress", srv.ep.getRecordByDomain)
	http.HandleFunc("/records/getDomains", srv.ep.getRecordByAddress)
	http.HandleFunc("/records/update", srv.ep.updateRecord)
	http.HandleFunc("/records/delete", srv.ep.deleteRecord)

	return srv
}

func (s *server) Start() error {
	return http.ListenAndServe(
		fmt.Sprintf(":%v", s.port),
		nil,
	)
}
