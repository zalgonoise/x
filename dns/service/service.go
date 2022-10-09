package service

import (
	"context"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/dns/uimpdns"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/store/uimpstore"
)

type Service interface {
	dns.DNSRepository
	store.StoreRepository
}

type service struct {
	dns   dns.DNSRepository
	store store.StoreRepository
}

func New(dns dns.DNSRepository, store store.StoreRepository) Service {
	if dns == nil {
		dns = uimpdns.New()
	}
	if store == nil {
		store = uimpstore.New()
	}

	return &service{
		dns:   dns,
		store: store,
	}
}

func (s *service) ParseQuery(m *dnsr.Msg) {
	s.dns.ParseQuery(m)
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

func (s *service) Add(ctx context.Context, r store.Record) error {
	return s.store.Add(ctx, r)
}

func (s *service) List(ctx context.Context) ([]store.Record, error) {
	return s.store.List(ctx)
}

func (s *service) GetByAddr(ctx context.Context, addr string) (store.Record, error) {
	return s.store.GetByAddr(ctx, addr)
}

func (s *service) GetByDest(ctx context.Context, addr string) ([]store.Record, error) {
	return s.store.GetByDest(ctx, addr)
}

func (s *service) Update(ctx context.Context, addr string, r store.Record) error {
	return s.store.Update(ctx, addr, r)
}

func (s *service) Delete(ctx context.Context, addr string) error {
	return s.store.Delete(ctx, addr)
}
