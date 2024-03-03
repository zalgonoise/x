package authz

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const authHeader = "Authorization"

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
			token  []byte
		)

		if len(values) > 0 {
			token = []byte(values[0])
		}

		if _, err := client.VerifyToken(ctx, &pb.AuthRequest{Token: token}); err != nil {
			return ctx, err
		}

		// TODO: inject ID data to context
		return ctx, nil
	}
}
