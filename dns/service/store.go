package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/zalgonoise/x/dns/store"
)

// AddRecord uses the store.Repository to create a DNS Record
func (s *service) AddRecord(ctx context.Context, r *store.Record) error {
	err := s.store.Create(ctx, r)
	if err != nil {
		return fmt.Errorf("couldn't add target record: %w", err)
	}
	return nil
}

// AddRecords uses the store.Repository to create a set of DNS Records
func (s *service) AddRecords(ctx context.Context, rs ...*store.Record) error {
	err := s.store.Create(ctx, rs...)
	if err != nil {
		return fmt.Errorf("couldn't add target records: %w", err)
	}
	return nil
}

// ListRecord uses the store.Repository to return all DNS Records
func (s *service) ListRecords(ctx context.Context) ([]*store.Record, error) {
	rs, err := s.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't list any records: %w", err)
	}
	if len(rs) == 0 {
		return rs, fmt.Errorf("no records in the store: %w", store.ErrZeroRecords)
	}
	return rs, nil
}

// GetRecordByDomain uses the store.Repository to return the DNS Record associated with
// the domain name and record type found in store.Record `r`
//
// Returns a NoName error if no domain name is provided
// Returns a NoType if no record type is provided
func (s *service) GetRecordByTypeAndDomain(ctx context.Context, rtype, domain string) (*store.Record, error) {
	if domain == "" {
		return nil, ErrNoName
	}
	if rtype == "" {
		return nil, ErrNoType
	}

	r, err := s.store.FilterByTypeAndDomain(ctx, rtype, domain)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch target record: %w", err)
	}
	return r, nil
}

// GetRecordByDomain uses the store.Repository to return the DNS Records associated with
// the IP address found in store.Record `r`
//
// Returns a NoAddr error if no IP address is provided
func (s *service) GetRecordByAddress(ctx context.Context, ip string) ([]*store.Record, error) {
	if ip == "" {
		return nil, ErrNoAddr
	}

	rs, err := s.store.FilterByDest(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch target records: %w", err)
	}
	if len(rs) == 0 {
		return rs, fmt.Errorf("no records in the store: %w", store.ErrZeroRecords)
	}
	return rs, nil
}

// UpdateRecord uses the store.Repository to update the record with domain name `domain`,
// based on the data provided in store.Record `r`
//
// Returns a NoAddr error if no IP address is provided
func (s *service) UpdateRecord(ctx context.Context, domain string, r *store.Record) error {
	if domain == "" {
		return store.ErrNoName
	}
	if r.Type == "" {
		return store.ErrNoType
	}
	// if record is nil or empty, request is parsed as a delete operation
	if r == nil || reflect.DeepEqual(r, &store.Record{}) {
		return s.store.Delete(ctx, &store.Record{Name: domain})
	}
	_, err := s.store.FilterByTypeAndDomain(ctx, r.Type, domain)
	if err != nil && errors.Is(err, store.ErrDoesNotExist) {
		return fmt.Errorf("couldn't find target record: %w", err)
	}

	err = s.store.Update(ctx, domain, r)
	if err != nil {
		return fmt.Errorf("couldn't update target record: %w", err)
	}
	return nil
}

// DeleteRecord uses the store.Repository to remove the store.Record based on input `r`
func (s *service) DeleteRecord(ctx context.Context, r *store.Record) error {
	err := s.store.Delete(ctx, r)
	if err != nil {
		return fmt.Errorf("couldn't delete target record: %w", err)
	}
	return nil
}
