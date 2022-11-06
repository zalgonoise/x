package httpapi

import (
	"errors"

	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/store"
)

var (
	ErrInvalidBody = errors.New("invalid body")
	ErrInvalidJSON = errors.New("body contains invalid JSON")
	ErrInternal    = errors.New("internal error")
)

type DNSResponse struct {
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type HealthResponse struct {
	Message string         `json:"message,omitempty"`
	Report  *health.Report `json:"report,omitempty"`
}

type StoreResponse struct {
	Success bool            `json:"success,omitempty"`
	Message string          `json:"message,omitempty"`
	Record  *store.Record   `json:"record,omitempty"`
	Records *[]store.Record `json:"records,omitempty"`
	Error   string          `json:"error,omitempty"`
}
