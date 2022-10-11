package service

import (
	"context"
	"errors"
	"reflect"

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
	dns.Repository
	store.Repository
}

type service struct {
	dns   dns.Repository
	store store.Repository
}

func New(dnsR dns.Repository, storeR store.Repository) *service {
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

func (s *service) Start() error {
	// route queries
	go func() {
		for m := range s.dns.Link() {
			ctx := context.Background()
			answer, err := s.store.GetByDomain(ctx, m)
			if err != nil {
				// logger
				s.dns.Link() <- store.New().Build()
			} else {
				s.dns.Link() <- answer
			}
		}
	}()

	return s.dns.Start()
}

func (s *service) Stop() error {
	return s.dns.Stop()
}

func (s *service) Reload() error {
	err := s.dns.Stop()
	if err != nil {
		return err
	}
	return s.Start()
}

// func (s *service) Store(store store.Repository) {
// 	s.dns.Store(s.store)
// }

func (s *service) Add(ctx context.Context, r ...*store.Record) error {
	return s.store.Add(ctx, r...)
}

func (s *service) List(ctx context.Context) ([]*store.Record, error) {
	return s.store.List(ctx)
}

func (s *service) GetByDomain(ctx context.Context, r *store.Record) (*store.Record, error) {
	if r.Name == "" {
		return nil, ErrNoName
	}
	if r.Type == "" {
		return nil, ErrNoType
	}

	return s.store.GetByDomain(ctx, r)
}

func (s *service) GetByDest(ctx context.Context, r *store.Record) ([]*store.Record, error) {
	if r.Addr == "" {
		return nil, ErrNoAddr
	}

	return s.store.GetByDest(ctx, r)
}

func (s *service) Update(ctx context.Context, addr string, r *store.Record) error {
	if addr == "" {
		return ErrNoAddr
	}
	// if record is nil or empty, request is parsed as a delete operation
	if r == nil || reflect.DeepEqual(r, &store.Record{}) {
		return s.store.Delete(ctx, &store.Record{Addr: addr})
	}

	return s.store.Update(ctx, addr, r)
}

func (s *service) Delete(ctx context.Context, r *store.Record) error {
	if r.Name == "" && r.Type == "" && r.Addr == "" {
		return ErrEmtpyRecord
	}
	return s.store.Delete(ctx, r)
}
