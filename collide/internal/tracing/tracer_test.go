package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestTracer(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		setup func(*testing.T) ShutdownFunc
	}{
		{
			name: "Success/WithInit",
			setup: func(t *testing.T) ShutdownFunc {
				done, err := Init(context.Background(), ServiceName, NoopExporter())
				require.NoError(t, err)

				return done
			},
		},
		{
			name: "Success/NoInit",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.setup != nil {
				done := testcase.setup(t)

				//nolint:errcheck // testing: we are sure noopTracer returns a nil error
				defer done(context.Background())
			}

			tracer := Tracer(ServiceName)
			require.NotNil(t, tracer)
		})
	}
}

func TestInit(t *testing.T) {
	for _, testcase := range []struct {
		name     string
		exporter sdktrace.SpanExporter
	}{
		{
			name:     "Success",
			exporter: NoopExporter(),
		},
		{
			name: "Success/NilExporter",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			ctx := context.Background()
			done, err := Init(ctx, ServiceName, testcase.exporter)
			//nolint:errcheck // testing: we are sure noopTracer returns a nil error
			defer done(ctx)
			require.NoError(t, err)
		})
	}
}
