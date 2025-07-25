package tracing

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zalgonoise/cfg"
)

const (
	authKey          = "Authorization"
	totalDialOptions = 2
)

// GRPCExporter creates a trace.SpanExporter using a gRPC connection to a tracing backend.
func GRPCExporter(
	ctx context.Context, url string, opts ...cfg.Option[Config],
) (sdktrace.SpanExporter, error) {
	config := cfg.Set(defaultConfig(), opts...)

	ctx, cancel := context.WithTimeout(ctx, config.timeout)
	defer cancel()

	dialOpts := make([]grpc.DialOption, 0, totalDialOptions)

	switch {
	case config.username != "" && config.password != "":
		dialOpts = append(dialOpts,
			// Disable "G402 (CWE-295): TLS MinVersion too low. (Confidence: HIGH, Severity: HIGH)":
			// Go has a minimum TLS version 1.2 set. By creating an empty tls.Config we're following that minimum version.
			//
			// To comply with this linter's rule, we'd need to add a minimum TLS version -- making the team revisit the code
			// on a future Go version where the minimum TLS version is updated (e.g. due to a crypto CVE), or making the app
			// less robust when preventing transport layer version downgrade attacks
			//
			// #nosec G402
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
			grpc.WithPerRPCCredentials(basicAuth{
				username: config.username,
				password: config.password,
			}),
		)
	default:
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(ctx, url, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	exporter, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return noopExporter{}, err
	}

	return exporter, nil
}

type basicAuth struct {
	username string
	password string
}

// GetRequestMetadata implements the credentials.PerRPCCredentials interface
//
// It returns a key-value (string) map of request headers used in basic authorization.
func (b basicAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		authKey: "Basic " + base64.StdEncoding.EncodeToString([]byte(b.username+":"+b.password)),
	}, nil
}

// RequireTransportSecurity implements the credentials.PerRPCCredentials interface.
func (basicAuth) RequireTransportSecurity() bool {
	return true
}

//nolint:revive // returning a private concrete type, but it is only usable internally
func NoopExporter() noopExporter {
	return noopExporter{}
}

type noopExporter struct{}

func (noopExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error {
	return nil
}

func (noopExporter) Shutdown(context.Context) error {
	return nil
}
