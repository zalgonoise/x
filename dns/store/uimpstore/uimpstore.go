package uimpstore

import (
	"context"
	"errors"

	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/store"
)

var (
	ErrUnimplementedStore error = errors.New("unimplemented DNS Record Store")
)

type UnimplementedStore struct{}

func (u UnimplementedStore) Add(context.Context, ...store.Record) error {
	return ErrUnimplementedStore
}

func (u UnimplementedStore) List(context.Context) ([]store.Record, error) {
	return nil, ErrUnimplementedStore
}

func (u UnimplementedStore) GetByAddr(context.Context, dns.RecordType, string) (store.Record, error) {
	return store.Record{}, ErrUnimplementedStore
}

func (u UnimplementedStore) GetByDest(context.Context, string) ([]store.Record, error) {
	return nil, ErrUnimplementedStore
}

func (u UnimplementedStore) Update(context.Context, string, store.Record) error {
	return ErrUnimplementedStore
}

func (u UnimplementedStore) Delete(context.Context, string) error {
	return ErrUnimplementedStore
}

func New() UnimplementedStore {
	return UnimplementedStore{}
}
