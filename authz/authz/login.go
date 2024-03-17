package authz

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha512"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/zalgonoise/x/authz/certs"
	"github.com/zalgonoise/x/authz/keygen"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"github.com/zalgonoise/x/authz/repository"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: Login must support creating up to two challenges; not just retrieve the first valid one
func (a *Authz) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.Login", trace.WithAttributes(
		attribute.String("service", req.Name),
		attribute.String("id.pub_key", string(req.Id.PublicKey)),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveServiceLoginLatency(ctx, req.Name, time.Since(start))
	}()

	a.metrics.IncServiceLoginRequests(req.Name)
	a.logger.DebugContext(ctx, "new login request", slog.String("service", req.Name))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.WarnContext(ctx, "invalid request", slog.Any("request", req))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	storedPEM, _, err := a.services.GetService(ctx, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
			a.metrics.IncServiceLoginFailed(req.Name)

			a.logger.WarnContext(ctx, "couldn't find service in the database", slog.String("service", req.Name))

			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.ErrorContext(ctx, "failed to fetch stored public key",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	servicePubKey, err := keygen.DecodePublic(req.Service.PublicKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.WarnContext(ctx, "invalid service public key PEM bytes",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidServicePublicKey.Error())
	}

	if !a.privateKey.PublicKey.Equal(servicePubKey) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.WarnContext(ctx, "mismatching service public keys",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidServicePublicKey.Error())
	}

	serviceCert, err := certs.Decode(req.Service.Certificate)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.WarnContext(ctx, "invalid service certificate bytes",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidServiceCertificate.Error())
	}

	if !a.cert.Equal(serviceCert) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.WarnContext(ctx, "mismatching service cert",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidServiceCertificate.Error())
	}

	pubKey, err := keygen.DecodePublic(req.Id.PublicKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.WarnContext(ctx, "invalid public key PEM bytes",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidIDPublicKey.Error())
	}

	cert, err := certs.Decode(req.Id.Certificate)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.WarnContext(ctx, "invalid cert bytes",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidIDCertificate.Error())
	}

	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.WarnContext(ctx, "failed to retrieve public key from certificate",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidIDCertificate.Error())
	}

	storedPub, err := keygen.DecodePublic(storedPEM)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.ErrorContext(ctx, "failed to decode stored public key",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if !pub.Equal(storedPub) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.WarnContext(ctx, "mismatching public keys",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.PermissionDenied, ErrInvalidIDPublicKey.Error())
	}

	if !pubKey.Equal(storedPub) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.WarnContext(ctx, "mismatching public keys",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.PermissionDenied, ErrInvalidIDPublicKey.Error())
	}

	if time.Now().After(cert.NotAfter) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.WarnContext(ctx, "expired certificate",
			slog.String("service", req.Name), slog.Time("expiry", cert.NotAfter),
			slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidIDCertificate.Error())
	}

	// check if there is a valid challenge to provide
	challenges, err := a.tokens.ListChallenges(ctx, req.Name)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.ErrorContext(ctx, "failed to search for challenges in the database",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(challenges) > 0 {
		return &pb.LoginResponse{Challenge: challenges[0].Raw, ExpiresOn: challenges[0].Expiry.UnixMilli()}, nil
	}

	// create a new challenge
	challenge, err := a.random.Random()
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.ErrorContext(ctx, "failed to generate challenge",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	expiry := time.Now().Add(a.challengeExpiry)

	if err = a.tokens.CreateChallenge(ctx, req.Name, challenge, expiry); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(req.Name)

		a.logger.ErrorContext(ctx, "failed to store challenge",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{Challenge: challenge, ExpiresOn: expiry.UnixMilli()}, nil
}

func (a *Authz) Token(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.Token", trace.WithAttributes(
		attribute.String("service", req.Name),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveServiceTokenLatency(ctx, req.Name, time.Since(start))
	}()

	a.metrics.IncServiceTokenRequests(req.Name)
	a.logger.DebugContext(ctx, "new token request",
		slog.String("service", req.Name))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(req.Name)

		a.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	storedPub, _, err := a.services.GetService(ctx, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
			a.metrics.IncServiceTokenFailed(req.Name)

			a.logger.WarnContext(ctx, "service does not exist",
				slog.String("service", req.Name), slog.String("error", err.Error()))

			return nil, status.Error(codes.InvalidArgument, ErrInvalidService.Error())
		}

		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(req.Name)

		a.logger.ErrorContext(ctx, "failed to get service details",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	pub, err := keygen.DecodePublic(storedPub)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(req.Name)

		a.logger.ErrorContext(ctx, "failed to decode public key",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	challenges, err := a.tokens.ListChallenges(ctx, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
			a.metrics.IncServiceTokenFailed(req.Name)

			a.logger.WarnContext(ctx, "couldn't find a challenge for this token request",
				slog.String("service", req.Name), slog.String("error", err.Error()))

			return nil, status.Error(codes.InvalidArgument, ErrInvalidChallenge.Error())
		}

		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(req.Name)

		a.logger.ErrorContext(ctx, "failed to get challenge",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	var (
		idx   int
		valid bool
	)

	for idx = range challenges {
		h := sha512.Sum512(challenges[idx].Raw)

		if ecdsa.VerifyASN1(pub, h[:], req.SignedChallenge) {
			valid = true

			break
		}
	}

	if !valid {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(req.Name)

		a.logger.WarnContext(ctx, "invalid signature",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidSignature.Error())
	}

	now := time.Now()

	tokens, err := a.tokens.ListTokens(ctx, req.Name)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(req.Name)

		a.logger.ErrorContext(ctx, "failed to lookup tokens for this service in the database",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(tokens) > 0 {
		if err = a.tokens.DeleteChallenge(ctx, req.Name, challenges[idx].Raw); err != nil {
			a.logger.WarnContext(ctx, "failed to delete challenge in the database",
				slog.String("service", req.Name), slog.String("error", err.Error()),
			)
		}

		return &pb.TokenResponse{
			Token:     string(tokens[0].Raw),
			ExpiresOn: tokens[0].Expiry.UnixMilli(),
		}, nil
	}

	exp := now.Add(a.tokenExpiry)

	token, err := keygen.NewToken(a.privateKey, a.name, exp, keygen.WithClaim(keygen.Claim{
		Service: req.Name,
		Authz:   a.name,
	}))
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(req.Name)

		a.logger.ErrorContext(ctx, "failed to generate JWT",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = a.tokens.CreateToken(ctx, req.Name, token, exp); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(req.Name)

		a.logger.ErrorContext(ctx, "failed to store token",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = a.tokens.DeleteChallenge(ctx, req.Name, challenges[idx].Raw); err != nil {
		a.logger.WarnContext(ctx, "failed to delete challenge in the database",
			slog.String("service", req.Name), slog.String("error", err.Error()),
		)
	}

	return &pb.TokenResponse{
		Token:     string(token),
		ExpiresOn: exp.UnixMilli(),
	}, nil
}

func (a *Authz) VerifyToken(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.VerifyToken")
	defer span.End()

	start := time.Now()

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenVerificationFailed("")
		a.logger.WarnContext(ctx, "invalid request", slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := keygen.ParseToken([]byte(req.Token), &a.privateKey.PublicKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenVerificationFailed("")
		a.logger.WarnContext(ctx, "failed to decode token")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	defer func() {
		a.metrics.ObserveServiceTokenVerificationLatency(ctx, token.Claim.Service, time.Since(start))
	}()

	a.metrics.IncServiceTokenVerifications(token.Claim.Service)
	a.logger.DebugContext(ctx, "new token verification request", slog.String("service", token.Claim.Service))

	tokens, err := a.tokens.ListTokens(ctx, token.Claim.Service)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenVerificationFailed(token.Claim.Service)
		a.logger.WarnContext(ctx, "invalid token")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	idx := slices.IndexFunc(tokens, func(token repository.Token) bool {
		return string(token.Raw) == req.Token
	})

	switch idx {
	case -1:
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenVerificationFailed(token.Claim.Service)
		a.logger.WarnContext(ctx, "invalid token")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	default:
		// expiry comes in GMT time (+00:00), truncated to the seconds; ensure comparison matches format
		if !tokens[idx].Expiry.Truncate(time.Second).Equal(token.Expiry.In(time.Local)) {
			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
			a.metrics.IncServiceTokenVerificationFailed(token.Claim.Service)
			a.logger.WarnContext(ctx, "invalid token")

			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return &pb.AuthResponse{}, nil
	}
}
