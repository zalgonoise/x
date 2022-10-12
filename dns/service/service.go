package service

import (
	"context"
	"errors"
	"reflect"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/dns"
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
}

type service struct {
	dns   dns.Repository
	store store.Repository
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

func (s *service) AddRecord(ctx context.Context, r *store.Record) error {
	return s.store.Add(ctx, r)
}

func (s *service) AddRecords(ctx context.Context, rs ...*store.Record) error {
	return s.store.Add(ctx, rs...)
}

func (s *service) ListRecords(ctx context.Context) ([]*store.Record, error) {
	return s.store.List(ctx)
}

func (s *service) GetRecordByDomain(ctx context.Context, r *store.Record) (*store.Record, error) {
	if r.Name == "" {
		return nil, ErrNoName
	}
	if r.Type == "" {
		return nil, ErrNoType
	}

	return s.store.GetByDomain(ctx, r)
}

func (s *service) GetRecordByAddress(ctx context.Context, ip string) ([]*store.Record, error) {
	if ip == "" {
		return nil, ErrNoAddr
	}

	return s.store.GetByDest(ctx, store.New().Addr(ip).Build())
}

func (s *service) UpdateRecord(ctx context.Context, domain string, r *store.Record) error {
	if domain == "" {
		return ErrNoAddr
	}
	// if record is nil or empty, request is parsed as a delete operation
	if r == nil || reflect.DeepEqual(r, &store.Record{}) {
		return s.store.Delete(ctx, &store.Record{Name: domain})
	}

	return s.store.Update(ctx, domain, r)
}

func (s *service) DeleteRecord(ctx context.Context, r *store.Record) error {
	return nil
}

func (s *service) AnswerDNS(r *store.Record, m *dnsr.Msg) {
	ctx := context.Background()
	answer, err := s.store.GetByDomain(ctx, r)
	if err != nil || answer.Addr == "" {
		return
	}
	s.dns.Answer(answer, m)
}
