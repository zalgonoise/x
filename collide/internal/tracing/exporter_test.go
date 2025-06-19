package tracing

import (
	"context"
	"github.com/zalgonoise/x/collide/internal/config"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGRPCExporter(t *testing.T) {
	for _, testcase := range []struct {
		name string
		cfg  config.Tracing
		err  error
	}{
		{
			name: "Success",
			cfg: config.Tracing{
				URI: "localhost:38088",
			},
		},
		{
			name: "Success/WithCredentials",
			cfg: config.Tracing{
				URI:      "localhost:38088",
				Username: "gopher",
				Password: "goroutine",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			_, err := GRPCExporter(testcase.cfg)
			if err != nil {
				require.ErrorIs(t, testcase.err, err)

				return
			}

			require.NoError(t, err)
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
