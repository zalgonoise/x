package service

import (
	"context"
	"reflect"

	"github.com/zalgonoise/x/dns/store"
)

func (s *service) AddRecord(ctx context.Context, r *store.Record) error {
	return s.store.Create(ctx, r)
}

func (s *service) AddRecords(ctx context.Context, rs ...*store.Record) error {
	return s.store.Create(ctx, rs...)
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

	return s.store.FilterByDomain(ctx, r)
}

func (s *service) GetRecordByAddress(ctx context.Context, ip string) ([]*store.Record, error) {
	if ip == "" {
		return nil, ErrNoAddr
	}

	return s.store.FilterByDest(ctx, store.New().Addr(ip).Build())
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
