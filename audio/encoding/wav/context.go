package wav

import (
	"context"

	"github.com/google/uuid"
)

type contextStreamID string

const (
	ContextStreamID contextStreamID = "stream_id"
)

func WithID(ctx context.Context) context.Context {
	if _, ok := GetID(ctx); ok {
		return ctx
	}

	return context.WithValue(ctx, ContextStreamID, uuid.New().String())
}

func GetID(ctx context.Context) (string, bool) {
	if idValue := ctx.Value(ContextStreamID); idValue != nil {
		if id, ok := idValue.(string); ok {
			return id, ok
		}
	}

	return "", false
}

func GetOrCreateID(ctx context.Context) string {
	if idValue := ctx.Value(ContextStreamID); idValue != nil {
		if id, ok := idValue.(string); ok {
			return id
		}
	}

	return uuid.New().String()
}
