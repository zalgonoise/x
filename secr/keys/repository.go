package keys

import "context"

// Repository describes the action exposed by the keys store
type Repository interface {
	// Get fetches the secret identified by `k` in the bucket `bucket`,
	// returning a slice of bytes for the value and an error
	Get(ctx context.Context, bucket, k string) ([]byte, error)
	// Set creates or overwrites a secret identified by `k` with value `v`, in
	// bucket `bucket`. Returns an error
	Set(ctx context.Context, bucket, k string, v []byte) error
	// Delete removes the secret identified by `k` in bucket `bucket`, returning an error
	Delete(ctx context.Context, bucket, k string) error
	// Purge removes all the secrets in the bucket `bucket`, returning an error
	Purge(ctx context.Context, bucket string) error
}
