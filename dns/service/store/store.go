package store

import "context"

type Record struct{}

type StoreRepository interface {
	AddRecord(ctx context.Context, r Record) error
	GetRecords(ctx context.Context) ([]Record, error)
	GetRecordByAddr(ctx context.Context, addr string) (Record, error)
	GetRecordByDest(ctx context.Context, addr string) (Record, error)
	UpdateRecord(ctx context.Context, addr string, r Record) error
	DeleteRecord(ctx context.Context, addr string) error
}
