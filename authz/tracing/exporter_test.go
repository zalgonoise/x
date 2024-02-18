package tracing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/cfg"
)

func TestGRPCExporter(t *testing.T) {
	testcases := []struct {
		name      string
		uri       string
		cfg       []cfg.Option[Config]
		errString string
	}{
		{
			name: "Success/ValidAddressButNoConnection",
			uri:  "localhost:38088",
			cfg: []cfg.Option[Config]{
				WithTimeout(250 * time.Millisecond),
			},
		},
		{
			name: "Success/ValidAddressButNoConnection/WithCredentials",
			uri:  "localhost:38088",
			cfg: []cfg.Option[Config]{
				WithTimeout(250 * time.Millisecond),
				WithCredentials("gopher", "goroutine"),
			},
		},
		{
			name: "Fail/NoAddress",
			cfg: []cfg.Option[Config]{
				WithTimeout(250 * time.Millisecond),
			},
			// from https://github.com/grpc/grpc-go/blob/master/internal/resolver/passthrough/passthrough.go#L35
			errString: "passthrough: received empty target in Build()",
		},
	}

	for i, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			_, err := GRPCExporter(context.Background(), testcase.uri, testcases[i].cfg...)
			if err != nil {
				require.NotEmpty(t, testcase.errString, err)
				require.Contains(t, err.Error(), testcase.errString)

				return
			}

			require.Empty(t, testcase.errString)
		})
	}
}

func TestBasicAuth_GetRequestMetadata(t *testing.T) {
	for _, testcase := range []struct {
		name     string
		username string
		password string
		wants    string
	}{
		{
			name:     "Simple",
			username: "gopher",
			password: "goroutine",
			wants:    "Basic Z29waGVyOmdvcm91dGluZQ==",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			headers, err := basicAuth{
				username: testcase.username,
				password: testcase.password,
			}.GetRequestMetadata(context.Background())

			require.NoError(t, err)
			value, ok := headers[authKey]
			require.True(t, ok)
			require.Equal(t, testcase.wants, value)
		})
	}
}

func TestBasicAuth_RequireTransportSecurity(t *testing.T) {
	require.Equal(t, true, basicAuth{}.RequireTransportSecurity())
}

func TestNoopExporter_ExportSpans(t *testing.T) {
	require.NoError(t, noopExporter{}.ExportSpans(context.Background(), nil))
}

func TestNoopExporter_Shutdown(t *testing.T) {
	require.NoError(t, noopExporter{}.Shutdown(context.Background()))
}
