package service

import (
	"context"
	"errors"
	"time"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/store"
)

var (
	ErrNoAddr      = errors.New("no IP address provided")
	ErrNoName      = errors.New("no domain name provided")
	ErrNoType      = errors.New("no DNS record type provided")
	ErrEmtpyRecord = errors.New("record cannot be empty")
)

type Service interface {
	// store-repository methods
	AddRecord(context.Context, *store.Record) error
	AddRecords(context.Context, ...*store.Record) error
	ListRecords(context.Context) ([]*store.Record, error)
	GetRecordByDomain(context.Context, *store.Record) (*store.Record, error)
	GetRecordByAddress(context.Context, string) ([]*store.Record, error)
	UpdateRecord(context.Context, string, *store.Record) error
	DeleteRecord(context.Context, *store.Record) error

	// dns-repository methods
	AnswerDNS(*store.Record, *dnsr.Msg)

	// health-repository methods
	StoreHealth(int, time.Duration) *health.Report
	DNSHealth() *health.Report
}

type Storing interface {
	AddRecord(context.Context, *store.Record) error
	AddRecords(context.Context, ...*store.Record) error
	ListRecords(context.Context) ([]*store.Record, error)
	GetRecordByDomain(context.Context, *store.Record) (*store.Record, error)
	GetRecordByAddress(context.Context, string) ([]*store.Record, error)
	UpdateRecord(context.Context, string, *store.Record) error
	DeleteRecord(context.Context, *store.Record) error
}

type Answering interface {
	GetRecordByDomain(context.Context, *store.Record) (*store.Record, error)
	AnswerDNS(*store.Record, *dnsr.Msg)
}

type service struct {
	dns    dns.Repository
	store  store.Repository
	health health.Repository
}

func New(dnsR dns.Repository, storeR store.Repository) Service {
	if dnsR == nil {
		dnsR = dns.Unimplemented()
	}
	if storeR == nil {
		storeR = store.Unimplemented()
	}

	return &service{
		dns:   dnsR,
		store: storeR,
	}
}
