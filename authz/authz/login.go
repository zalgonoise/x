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

	exit := withExit[pb.LoginRequest, pb.LoginResponse](
		ctx, a.logger, req, func() { a.metrics.IncServiceLoginFailed(req.Name) }, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	storedPEM, _, err := a.services.GetService(ctx, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exit(codes.InvalidArgument, "couldn't find service in the database", ErrInvalidService)
		}

		return exit(codes.Internal, "failed to fetch stored public key", err)
	}

	servicePubKey, err := keygen.DecodePublic(req.Service.PublicKey)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid service public key PEM bytes", err)
	}

	if !a.privateKey.PublicKey.Equal(servicePubKey) {
		return exit(codes.InvalidArgument, "mismatching service public keys", ErrInvalidServicePublicKey)
	}

	serviceCert, err := certs.Decode(req.Service.Certificate)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid service cert bytes", err)
	}

	if !a.cert.Equal(serviceCert) {
		return exit(codes.InvalidArgument, "mismatching service cert", ErrInvalidServiceCertificate)
	}

	pubKey, err := keygen.DecodePublic(req.Id.PublicKey)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid public key PEM bytes", err)
	}

	cert, err := certs.Decode(req.Id.Certificate)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid cert bytes", err)
	}

	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return exit(codes.InvalidArgument, "failed to retrieve public key from certificate", ErrInvalidIDCertificate)
	}

	storedPub, err := keygen.DecodePublic(storedPEM)
	if err != nil {
		return exit(codes.Internal, "failed to decode stored public key", err)
	}

	if !pub.Equal(storedPub) {
		return exit(codes.PermissionDenied, "mismatching public keys", ErrInvalidIDPublicKey)
	}

	if !pubKey.Equal(storedPub) {
		return exit(codes.PermissionDenied, "mismatching public keys", ErrInvalidIDPublicKey)
	}

	if time.Now().After(cert.NotAfter) {
		a.logger.DebugContext(ctx, "expired certificate",
			slog.Time("expiry", cert.NotAfter), slog.String("service", req.Name))

		return exit(codes.InvalidArgument, "expired certificate", ErrInvalidIDCertificate)
	}

	// check if there is a valid challenge to provide
	challenge, expiry, err := a.tokens.GetChallenge(ctx, req.Name)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return exit(codes.Internal, "failed to search for challenges in the database", err)
	}

	if err == nil && len(challenge) > 0 {
		if expiry.After(start) {
			return &pb.LoginResponse{Challenge: challenge, ExpiresOn: expiry.UnixMilli()}, nil
		}

		if err = a.tokens.DeleteChallenge(ctx, req.Name); err != nil {
			return exit(codes.Internal, "failed to remove expired challenge", err)
		}
	}

	// create a new challenge
	challenge, err = a.random.Random()
	if err != nil {
		return exit(codes.Internal, "failed to generate challenge", err)
	}

	expiry = time.Now().Add(a.challengeExpiry)

	if err = a.tokens.CreateChallenge(ctx, req.Name, challenge, expiry); err != nil {
		return exit(codes.Internal, "failed to store challenge", err)
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

	exit := withExit[pb.TokenRequest, pb.TokenResponse](
		ctx, a.logger, req, func() { a.metrics.IncServiceTokenFailed(req.Name) }, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	storedPub, _, err := a.services.GetService(ctx, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exit(codes.InvalidArgument, "service does not exist", err)
		}

		return exit(codes.Internal, "failed to get service details", err)
	}

	challenge, expiry, err := a.tokens.GetChallenge(ctx, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exit(codes.InvalidArgument, "couldn't find a challenge for this token request", err)
		}

		return exit(codes.Internal, "failed to get challenge", err)
	}

	if time.Now().After(expiry) {
		return exit(codes.InvalidArgument, "challenge is expired", ErrExpiredChallenge)
	}

	pub, err := keygen.DecodePublic(storedPub)
	if err != nil {
		return exit(codes.Internal, "failed to decode public key", err)
	}

	h := sha512.Sum512(challenge)

	if !ecdsa.VerifyASN1(pub, h[:], req.SignedChallenge) {
		return exit(codes.InvalidArgument, "invalid signature", ErrInvalidSignature)
	}

	now := time.Now()

	tokens, err := a.tokens.ListTokens(ctx, req.Name)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return exit(codes.Internal, "failed to lookup tokens for this service in the database", err)
	}

	if len(tokens) > 0 {
		slices.SortFunc(tokens, func(a, b repository.Token) int {
			return b.Expiry.Compare(a.Expiry)
		})

		if err := a.tokens.DeleteChallenge(ctx, req.Name); err != nil {
			a.logger.ErrorContext(ctx, "failed to delete challenge in the database",
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
		return exit(codes.Internal, "failed to generate JWT", err)
	}

	if err = a.tokens.CreateToken(ctx, req.Name, token, exp); err != nil {
		return exit(codes.Internal, "failed to store token", err)
	}

	if err := a.tokens.DeleteChallenge(ctx, req.Name); err != nil {
		a.logger.ErrorContext(ctx, "failed to delete challenge in the database",
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

	token, err := keygen.ParseToken([]byte(req.Token), &a.privateKey.PublicKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenVerificationFailed("")
		a.logger.WarnContext(ctx, "failed to decode token")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	start := time.Now()
	defer func() {
		a.metrics.ObserveServiceTokenVerificationLatency(ctx, token.Claim.Service, time.Since(start))
	}()

	a.metrics.IncServiceTokenVerifications(token.Claim.Service)
	a.logger.DebugContext(ctx, "new token verification request")

	exit := withExit[pb.AuthRequest, pb.AuthResponse](
		ctx, a.logger, req, func() {
			a.metrics.IncServiceTokenVerificationFailed(token.Claim.Service)
		}, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	tokens, err := a.tokens.ListTokens(ctx, token.Claim.Service)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid token", err)
	}

	idx := slices.IndexFunc(tokens, func(token repository.Token) bool {
		return string(token.Raw) == req.Token
	})

	switch idx {
	case -1:
		return exit(codes.InvalidArgument, "invalid token", err)
	default:
		// expiry comes in GMT time (+00:00), truncated to the seconds; ensure comparison matches format
		if !tokens[idx].Expiry.Truncate(time.Second).Equal(token.Expiry.In(time.Local)) {
			return exit(codes.InvalidArgument, "invalid token", err)
		}

		return &pb.AuthResponse{}, nil
	}
}
