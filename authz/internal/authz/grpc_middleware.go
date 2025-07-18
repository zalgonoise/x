package authz

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/zalgonoise/x/authz/internal/keygen"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
)

type contextKey string

const (
	authHeader            = "Authorization"
	ContextKey contextKey = "authz"
)

var (
	ErrEmptyHeaders = errors.New("empty headers")
)

func UnaryInterceptor(client pb.AuthzClient) grpc.UnaryServerInterceptor {
	return auth.UnaryServerInterceptor(authzFunc(client))
}

func StreamInterceptor(client pb.AuthzClient) grpc.StreamServerInterceptor {
	return auth.StreamServerInterceptor(authzFunc(client))
}

func authzFunc(client pb.AuthzClient) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {

			return ctx, ErrEmptyHeaders
		}

		var (
			values = md.Get(authHeader)
			token  string
		)

		if len(values) > 0 {
			token = values[0]
		}

		if _, err := client.VerifyToken(ctx, &pb.AuthRequest{Token: token}); err != nil {
			return ctx, err
		}

		t, err := keygen.ParseToken([]byte(token), nil)
		if err != nil {
			return ctx, err
		}

		return context.WithValue(ctx, ContextKey, t.Claim), nil
	}
}
