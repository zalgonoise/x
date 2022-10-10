package service

import (
	"context"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/store"
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

	// link DNS server to DNS Records store
	dnsR.Store(storeR)

	return &service{
		dns:   dnsR,
		store: storeR,
	}
}

func (s *service) HandleRequest(w dnsr.ResponseWriter, r *dnsr.Msg) {
	s.dns.HandleRequest(w, r)
}

func (s *service) Start() error {
	return s.dns.Start()
}

func (s *service) Stop() error {
	return s.dns.Stop()
}

func (s *service) Reload() error {
	return s.dns.Reload()
}

func (s *service) Store(store store.Repository) {
	s.dns.Store(s.store)
}

func (s *service) Add(ctx context.Context, r ...*store.Record) error {
	return s.store.Add(ctx, r...)
}

func (s *service) List(ctx context.Context) ([]*store.Record, error) {
	return s.store.List(ctx)
}

func (s *service) GetByDomain(ctx context.Context, r *store.Record) (*store.Record, error) {
	return s.store.GetByDomain(ctx, r)
}

func (s *service) GetByDest(ctx context.Context, r *store.Record) ([]*store.Record, error) {
	return s.store.GetByDest(ctx, r)
}

func (s *service) Update(ctx context.Context, addr string, r *store.Record) error {
	return s.store.Update(ctx, addr, r)
}

func (s *service) Delete(ctx context.Context, r *store.Record) error {
	return s.store.Delete(ctx, r)
}
