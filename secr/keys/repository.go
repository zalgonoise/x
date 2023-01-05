package keys

import "context"

type Repository interface {
	Get(ctx context.Context, bucket, k string) ([]byte, error)
	Set(ctx context.Context, bucket, k string, v []byte) error
	Delete(ctx context.Context, bucket, k string) error
	Purge(ctx context.Context, bucket string) error
}
