package gacha

import (
	"context"
	"uuid"
)

type idKey struct{}

// ContextWithID returns a copy of ctx carrying the ID of the entry being handled. The poller calls it before running
// a Handler. It is also useful to test handlers that read the ID.
func ContextWithID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, idKey{}, id)
}

// IDFromContext returns the ID of the entry being handled, as set by ContextWithID. Handlers can use it as an
// idempotency key.
func IDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(idKey{}).(uuid.UUID)

	return id, ok
}
