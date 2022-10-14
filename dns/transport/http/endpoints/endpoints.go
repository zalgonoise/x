package endpoints

import (
	"net/http"

	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/transport/udp"
)

type endpoints struct {
	s   service.Service
	udp udp.Server
}

type HTTPAPI interface {
	startDNS(w http.ResponseWriter, r *http.Request)
	stopDNS(w http.ResponseWriter, r *http.Request)
	reloadDNS(w http.ResponseWriter, r *http.Request)

	addRecord(w http.ResponseWriter, r *http.Request)
	listRecords(w http.ResponseWriter, r *http.Request)
	getRecordByDomain(w http.ResponseWriter, r *http.Request)
	getRecordByAddress(w http.ResponseWriter, r *http.Request)
	updateRecord(w http.ResponseWriter, r *http.Request)
	deleteRecord(w http.ResponseWriter, r *http.Request)
}

func NewAPI(s service.Service, udps udp.Server) HTTPAPI {
	return &endpoints{
		s:   s,
		udp: udps,
	}
}
