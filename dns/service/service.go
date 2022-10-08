package service

import (
	"github.com/zalgonoise/x/dns/service/dns"
	"github.com/zalgonoise/x/dns/service/store"
)

type Service interface {
	DNS() *dns.DNSRepository
	Store() *store.StoreRepository
}

type service struct {
	dns   *dns.DNSRepository
	store *store.StoreRepository
}

func New(dns *dns.DNSRepository, store *store.StoreRepository) Service {
	return &service{
		dns:   dns,
		store: store,
	}
}

func (s *service) DNS() *dns.DNSRepository {
	return s.dns
}

func (s *service) Store() *store.StoreRepository {
	return s.store
}
