package ca

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509/pkix"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/zalgonoise/x/authz/certs"
	"github.com/zalgonoise/x/authz/config"
	"github.com/zalgonoise/x/authz/database"
	"github.com/zalgonoise/x/authz/keygen"
	"github.com/zalgonoise/x/authz/log"
	"github.com/zalgonoise/x/authz/metrics"
	"github.com/zalgonoise/x/authz/repository"
	"github.com/zalgonoise/x/authz/tracing"
)

func TestCertificateAuthority(t *testing.T) {
	// cleanup before and after running the tests
	require.NoError(t, cleanup())
	defer func() {
		require.NoError(t, cleanup())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	logger := log.New("debug")

	service, done, err := initCA(ctx, logger)
	require.NoError(t, err)
	defer func() {
		done(t)
	}()

	t.Run("RegisterService", func(t *testing.T) {
		// sample keys
		key, err := keygen.New()
		require.NoError(t, err)

		privPEM, err := keygen.EncodePrivate(key)
		require.NoError(t, err)

		pubPEM, err := keygen.EncodePublic(&key.PublicKey)
		require.NoError(t, err)

		otherKey, err := keygen.New()
		require.NoError(t, err)

		otherPubPEM, err := keygen.EncodePublic(&otherKey.PublicKey)
		require.NoError(t, err)

		for _, testcase := range []struct {
			name  string
			req   *pb.CertificateRequest
			fails bool
		}{
			{
				name: "Success/Simple",
				req: &pb.CertificateRequest{
					Service:   "test.register.simple",
					PublicKey: pubPEM,
				},
			},
			{
				name: "Success/SecondRun",
				req: &pb.CertificateRequest{
					Service:   "test.register.simple",
					PublicKey: pubPEM,
				},
			},
			{
				name: "Fail/PrivateKeysInsteadOfPublic",
				req: &pb.CertificateRequest{
					Service:   "test.register.fails.priv_key",
					PublicKey: privPEM,
				},
				fails: true,
			},
			{
				name: "Fail/MissingServiceName",
				req: &pb.CertificateRequest{
					Service:   "",
					PublicKey: pubPEM,
				},
				fails: true,
			},
			{
				name: "Fail/InvalidPublicKey",
				req: &pb.CertificateRequest{
					Service:   "test.register.simple",
					PublicKey: otherPubPEM,
				},
				fails: true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				res, err := service.RegisterService(ctx, testcase.req)
				if err != nil {
					require.True(t, testcase.fails)

					return
				}

				require.NotNil(t, res.Certificate)
				require.NotZero(t, res.ExpiresOn)
			})
		}
	})

	t.Run("RoundTrip", func(t *testing.T) {
		// sample keys
		key, err := keygen.New()
		require.NoError(t, err)

		pubPEM, err := keygen.EncodePublic(&key.PublicKey)
		require.NoError(t, err)

		for _, testcase := range []struct {
			name  string
			req   *pb.CertificateRequest
			fails bool
		}{
			{
				name: "Success/Simple",
				req: &pb.CertificateRequest{
					Service:   "test.roundtrip.simple",
					PublicKey: pubPEM,
				},
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				registerRes, err := service.RegisterService(ctx, testcase.req)
				if err != nil {
					require.True(t, testcase.fails)

					return
				}

				require.NotNil(t, registerRes.Certificate)
				require.NotZero(t, registerRes.ExpiresOn)

				listRes, err := service.ListCertificates(ctx, testcase.req)
				if err != nil {
					require.True(t, testcase.fails)

					return
				}

				require.Len(t, listRes.Certificates, 1)
				require.Equal(t, registerRes.Certificate, listRes.Certificates[0].Certificate)
				require.Equal(t, registerRes.ExpiresOn, listRes.Certificates[0].ExpiresOn)

				_, err = service.VerifyCertificate(ctx, &pb.VerificationRequest{
					Service:     testcase.req.Service,
					Certificate: registerRes.Certificate,
				})
				if err != nil {
					require.True(t, testcase.fails)

					return
				}

				_, err = service.DeleteCertificate(ctx, &pb.CertificateDeletionRequest{
					Service:     testcase.req.Service,
					PublicKey:   testcase.req.PublicKey,
					Certificate: registerRes.Certificate,
				})
				if err != nil {
					require.True(t, testcase.fails)

					return
				}

				_, err = service.DeleteService(ctx, &pb.DeletionRequest{
					Service:   testcase.req.Service,
					PublicKey: testcase.req.PublicKey,
				})
				if err != nil {
					require.True(t, testcase.fails)

					return
				}

			})
		}
	})
}

func cleanup() error {
	if err := os.Remove("./testdata/ca.db"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func getKey(path string) (*ecdsa.PrivateKey, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	pem, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return keygen.DecodePrivate(pem)
}

func initCA(ctx context.Context, logger *slog.Logger) (*CertificateAuthority, func(t *testing.T), error) {
	m := metrics.NewMetrics()
	tracer := noop.NewTracerProvider().Tracer(tracing.ServiceNameCA)
	tracingDone := func(context.Context) error { return nil }

	conf := config.Config{
		HTTPPort: 8080,
		GRPCPort: 8081,
		Name:     "authz.certificate_authority",
		CA: config.CA{
			CertDurMonths: 12,
		},
		Database: config.Database{
			URI:             "./testdata/ca.db",
			CleanupTimeout:  5 * time.Minute,
			CleanupSchedule: "0 6 * * *",
		},
	}

	privateKey, err := getKey("./testdata/testkey_private.pem")
	if err != nil {
		return nil, nil, err
	}

	db, err := database.Open(conf.Database.URI)
	if err != nil {
		return nil, nil, err
	}

	if err = database.Migrate(ctx, db, database.CAService, logger); err != nil {
		return nil, nil, err
	}

	repo, err := repository.NewService(db,
		repository.WithLogger(logger),
		repository.WithMetrics(m),
		repository.WithTrace(tracer),
		repository.WithCleanupTimeout(conf.Database.CleanupTimeout),
		repository.WithCleanupSchedule(conf.Database.CleanupSchedule),
	)
	if err != nil {
		return nil, nil, err
	}

	caService, err := NewCertificateAuthority(
		repo,
		privateKey,
		WithLogger(logger),
		WithMetrics(m),
		WithTracer(tracer),
		WithTemplate(
			certs.WithName(pkix.Name{CommonName: conf.Name}),
			certs.WithDurMonth(conf.CA.CertDurMonths),
		),
	)
	if err != nil {
		_ = db.Close()

		return nil, nil, err
	}

	done := func(t *testing.T) {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		require.NoError(t, tracingDone(shutdownCtx))

		require.NoError(t, db.Close())
	}

	return caService, done, nil
}
