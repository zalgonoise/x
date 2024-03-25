package authz

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha512"
	"crypto/x509"
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

const (
	challengeLimit = 2
	tokenLimit     = 10
)

func (a *Authz) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	idCert, service, err := a.validateCertificate(ctx, req.IdCertificate)
	if err != nil {
		a.metrics.IncServiceLoginRequests(service)
		a.metrics.IncServiceLoginFailed(service)
		a.logger.WarnContext(ctx, "invalid request", slog.Any("request", req))

		if errors.Is(err, ErrInvalidPublicKey) {
			a.logger.WarnContext(ctx, "invalid ID public key",
				slog.String("service", service), slog.String("error", err.Error()))

			return nil, status.Error(codes.InvalidArgument, ErrInvalidIDPublicKey.Error())
		}

		a.logger.ErrorContext(ctx, "error matching public keys",
			slog.String("service", service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx, span := a.tracer.Start(ctx, "Authz.Login", trace.WithAttributes(
		attribute.String("service", service),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveServiceLoginLatency(ctx, service, time.Since(start))
	}()

	a.metrics.IncServiceLoginRequests(service)
	a.logger.DebugContext(ctx, "new login request", slog.String("service", service))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(service)

		a.logger.WarnContext(ctx, "invalid request", slog.Any("request", req))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// validate authz certificate
	serviceCert, err := certs.Decode(req.ServiceCertificate)
	if err != nil {
		// TODO: err invalid service cert
	}

	if !a.cert.Equal(serviceCert) {
		// TODO : err invalid service cert
	}

	// validate client certificate
	if err := certs.Verify(req.IdCertificate, a.root, a.intermediates); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(service)

		a.logger.WarnContext(ctx, "invalid ID certificate",
			slog.String("service", service), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidIDCertificate.Error())
	}

	if time.Now().After(idCert.NotAfter) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(service)

		a.logger.WarnContext(ctx, "expired certificate",
			slog.String("service", service), slog.Time("expiry", idCert.NotAfter),
			slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidIDCertificate.Error())
	}

	// check if there is a valid challenge to provide
	challenges, err := a.tokens.ListChallenges(ctx, service)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(service)

		a.logger.ErrorContext(ctx, "failed to search for challenges in the database",
			slog.String("service", service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	// allow up to 2 challenges to exist simultaneously
	if len(challenges) >= challengeLimit {
		return challenges[0], nil
	}

	// create a new challenge
	challenge, err := a.random.Random()
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(service)

		a.logger.ErrorContext(ctx, "failed to generate challenge",
			slog.String("service", service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	expiry := time.Now().Add(a.challengeExpiry)

	if err = a.tokens.CreateChallenge(ctx, service, challenge, expiry); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceLoginFailed(service)

		a.logger.ErrorContext(ctx, "failed to store challenge",
			slog.String("service", service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{Challenge: challenge, ExpiresOn: expiry.UnixMilli()}, nil
}

func (a *Authz) Token(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	_, service, err := a.validateCertificate(ctx, req.Certificate)
	if err != nil {
		a.metrics.IncServiceTokenRequests(service)
		a.metrics.IncServiceTokenFailed(service)
		a.logger.WarnContext(ctx, "invalid request", slog.Any("request", req))

		if errors.Is(err, ErrInvalidPublicKey) {
			a.logger.WarnContext(ctx, "invalid ID public key",
				slog.String("service", service), slog.String("error", err.Error()))

			return nil, status.Error(codes.InvalidArgument, ErrInvalidIDPublicKey.Error())
		}

		a.logger.ErrorContext(ctx, "error matching public keys",
			slog.String("service", service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx, span := a.tracer.Start(ctx, "Authz.Token", trace.WithAttributes(
		attribute.String("service", service),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveServiceTokenLatency(ctx, service, time.Since(start))
	}()

	a.metrics.IncServiceTokenRequests(service)
	a.logger.DebugContext(ctx, "new token request",
		slog.String("service", service))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(service)

		a.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	challenges, err := a.tokens.ListChallenges(ctx, service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
			a.metrics.IncServiceTokenFailed(service)

			a.logger.WarnContext(ctx, "couldn't find a challenge for this token request",
				slog.String("service", service), slog.String("error", err.Error()))

			return nil, status.Error(codes.InvalidArgument, ErrInvalidChallenge.Error())
		}

		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(service)

		a.logger.ErrorContext(ctx, "failed to get challenge",
			slog.String("service", service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	idx, err := a.verifyChallenge(ctx, service, req.SignedChallenge, challenges)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(service)

		if errors.Is(err, ErrInvalidChallenge) {
			a.logger.WarnContext(ctx, "couldn't match a challenge for this signed response",
				slog.String("service", service), slog.String("error", err.Error()))

			return nil, status.Error(codes.InvalidArgument, ErrInvalidChallenge.Error())
		}

		a.logger.ErrorContext(ctx, "failed to verify signature",
			slog.String("service", service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	tokens, err := a.tokens.ListTokens(ctx, service)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(service)

		a.logger.ErrorContext(ctx, "failed to lookup tokens for this service in the database",
			slog.String("service", service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	// allow up to 10 tokens to exist simultaneously
	if len(tokens) >= tokenLimit {
		if err = a.tokens.DeleteChallenge(ctx, service, challenges[idx].Challenge); err != nil {
			a.logger.WarnContext(ctx, "failed to delete challenge in the database",
				slog.String("service", service), slog.String("error", err.Error()),
			)
		}

		return tokens[0], nil
	}

	token, err := a.newToken(ctx, service, start)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenFailed(service)

		a.logger.ErrorContext(ctx, "failed to generate JWT",
			slog.String("service", service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return token, nil
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

	idx := slices.IndexFunc(tokens, func(response *pb.TokenResponse) bool {
		return response.Token == req.Token
	})

	switch idx {
	case -1:
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenVerificationFailed(token.Claim.Service)
		a.logger.WarnContext(ctx, "invalid token")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	default:
		if token.Expiry.Before(start) {
			span.SetStatus(otelcodes.Error, ErrExpiredToken.Error())
			span.RecordError(ErrExpiredToken)
			a.metrics.IncServiceTokenVerificationFailed(token.Claim.Service)
			a.logger.WarnContext(ctx, "expired token")

			return nil, status.Error(codes.InvalidArgument, ErrExpiredToken.Error())
		}

		return &pb.AuthResponse{}, nil
	}
}

func (a *Authz) validateCertificate(ctx context.Context, raw []byte) (*x509.Certificate, string, error) {
	cert, err := certs.Decode(raw)
	if err != nil {
		return nil, "", err
	}

	pubKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, cert.Subject.CommonName, ErrInvalidIDPublicKey
	}

	// validate client ID
	if err := a.validatePublicKeys(ctx, cert.Subject.CommonName, pubKey); err != nil {
		return nil, cert.Subject.CommonName, err
	}

	return cert, cert.Subject.CommonName, nil
}

func (a *Authz) verifyChallenge(ctx context.Context, name string, signed []byte, challenges []*pb.LoginResponse) (int, error) {
	storedPub, err := a.services.GetService(ctx, name)
	if err != nil {
		return -1, err
	}

	pub, err := keygen.DecodePublic(storedPub)
	if err != nil {
		return -1, err
	}

	var (
		idx   int
		valid bool
	)

	for idx = range challenges {
		h := sha512.Sum512(challenges[idx].Challenge)

		if ecdsa.VerifyASN1(pub, h[:], signed) {
			valid = true

			break
		}
	}

	if !valid {
		return -1, ErrInvalidChallenge
	}

	return idx, nil
}

func (a *Authz) newToken(ctx context.Context, name string, start time.Time) (*pb.TokenResponse, error) {
	exp := start.Add(a.tokenExpiry)

	token, err := keygen.NewToken(a.privateKey, a.name, exp, keygen.WithClaim(keygen.Claim{
		Service: name,
		Authz:   a.name,
	}))
	if err != nil {
		return nil, err
	}

	if err = a.tokens.CreateToken(ctx, name, token, exp); err != nil {
		return nil, err
	}

	return &pb.TokenResponse{
		Token:     string(token),
		ExpiresOn: exp.UnixMilli(),
	}, nil
}
