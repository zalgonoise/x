package keys

import (
	"context"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
)

type withTrace struct {
	r Repository
}

func WithTrace(r Repository) Repository {
	return withTrace{
		r: r,
	}
}

// Set creates or overwrites a secret identified by `k` with value `v`, in
// bucket `bucket`. Returns an error
func (t withTrace) Set(ctx context.Context, bucket, k string, v []byte) error {
	ctx, s := spanner.Start(ctx, "keys.Set")
	defer s.End()
	s.Add(
		attr.String("in_bucket", bucket),
	)

	err := t.r.Set(ctx, bucket, k, v)
	if err != nil {
		s.Event("error setting key", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// Get fetches the secret identified by `k` in the bucket `bucket`,
// returning a slice of bytes for the value and an error
func (t withTrace) Get(ctx context.Context, bucket, k string) ([]byte, error) {
	ctx, s := spanner.Start(ctx, "keys.Get")
	defer s.End()
	s.Add(
		attr.String("in_bucket", bucket),
	)

	value, err := t.r.Get(ctx, bucket, k)
	if err != nil {
		s.Event("error fetching key", attr.New("error", err.Error()))
		return value, err
	}
	return value, nil
}

// Delete removes the secret identified by `k` in bucket `bucket`, returning an error
func (t withTrace) Delete(ctx context.Context, bucket, k string) error {
	ctx, s := spanner.Start(ctx, "keys.Delete")
	defer s.End()
	s.Add(
		attr.String("in_bucket", bucket),
	)

	err := t.r.Delete(ctx, bucket, k)
	if err != nil {
		s.Event("error deleting key", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// Purge removes all the secrets in the bucket `bucket`, returning an error
func (t withTrace) Purge(ctx context.Context, bucket string) error {
	ctx, s := spanner.Start(ctx, "keys.Purge")
	defer s.End()
	s.Add(
		attr.String("in_bucket", bucket),
	)

	err := t.r.Purge(ctx, bucket)
	if err != nil {
		s.Event("error deleting bucket", attr.New("error", err.Error()))
		return err
	}
	return nil
}
