package endpoints

import (
	"net/http"

	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/transport/udp"
)

type endpoints struct {
	s   service.Storing
	udp udp.Server
}

type HTTPAPI interface {
	StartDNS(w http.ResponseWriter, r *http.Request)
	StopDNS(w http.ResponseWriter, r *http.Request)
	ReloadDNS(w http.ResponseWriter, r *http.Request)

	AddRecord(w http.ResponseWriter, r *http.Request)
	ListRecords(w http.ResponseWriter, r *http.Request)
	GetRecordByDomain(w http.ResponseWriter, r *http.Request)
	GetRecordByAddress(w http.ResponseWriter, r *http.Request)
	UpdateRecord(w http.ResponseWriter, r *http.Request)
	DeleteRecord(w http.ResponseWriter, r *http.Request)
}

func NewAPI(s service.Service, udps udp.Server) HTTPAPI {
	return &endpoints{
		s:   s,
		udp: udps,
	}
}
