package service

import (
	"context"
	"errors"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/health/simplehealth"
	"github.com/zalgonoise/x/dns/store"
)

var (
	ErrNoAddr      = errors.New("no IP address provided")
	ErrNoName      = errors.New("no domain name provided")
	ErrNoType      = errors.New("no DNS record type provided")
	ErrEmtpyRecord = errors.New("record cannot be empty")
)

type Service interface {
	StoreService
	DNSService
	HealthService
}

type StoreService interface {
	// store-repository methods
	AddRecord(context.Context, *store.Record) error
	AddRecords(context.Context, ...*store.Record) error
	ListRecords(context.Context) ([]*store.Record, error)
	GetRecordByDomain(context.Context, *store.Record) (*store.Record, error)
	GetRecordByAddress(context.Context, string) ([]*store.Record, error)
	UpdateRecord(context.Context, string, *store.Record) error
	DeleteRecord(context.Context, *store.Record) error
}

type DNSService interface {
	// dns-repository methods
	AnswerDNS(*store.Record, *dnsr.Msg)
}

type HealthService interface {
	// health-repository methods
	StoreHealth() *health.StoreReport
	DNSHealth() *health.DNSReport
	HTTPHealth() *health.HTTPReport
	Health() *health.Report
}

type StoreWithHealth interface {
	StoreService
	HealthService
}

type Answering interface {
	GetRecordByDomain(context.Context, *store.Record) (*store.Record, error)
	AnswerDNS(*store.Record, *dnsr.Msg)
}

type service struct {
	dns    dns.Repository
	store  store.Repository
	health health.Repository
	conf   *config.Config
}

func New(
	dnsR dns.Repository,
	storeR store.Repository,
	healthR health.Repository,
	conf *config.Config,
) Service {
	if dnsR == nil {
		dnsR = dns.Unimplemented()
	}
	if storeR == nil {
		storeR = store.Unimplemented()
	}
	if healthR == nil {
		healthR = simplehealth.New()
	}

	return &service{
		dns:    dnsR,
		store:  storeR,
		health: healthR,
		conf:   conf,
	}
}
