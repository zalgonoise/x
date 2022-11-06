package httpapi

import "net/http"

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

	Health(w http.ResponseWriter, r *http.Request)
}
