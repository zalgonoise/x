package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		env   map[string]string
		wants Config
	}{
		{
			name: "Default",
			wants: Config{
				HTTP: HTTP{
					Port:     8080,
					GRPCPort: 8081,
				},
				Tracks: Tracks{
					Path: "",
				},
				Logging: Logging{
					Level:      "INFO",
					WithSource: true,
					WithSpanID: true,
				},
				Tracing: Tracing{
					URI:      "",
					Username: "",
					Password: "",
				},
			},
		},
		{
			name: "Custom",
			env: map[string]string{
				"COLLIDE_HTTP_PORT":        "8088",
				"COLLIDE_GRPC_PORT":        "8089",
				"COLLIDE_TRACKS_PATH":      "tracks.yaml",
				"COLLIDE_LOG_LEVEL":        "DEBUG",
				"COLLIDE_LOG_WITH_SOURCE":  "false",
				"COLLIDE_LOG_WITH_SPAN_ID": "false",
				"COLLIDE_TRACING_URI":      "localhost:8000",
				"COLLIDE_TRACING_USERNAME": "user",
				"COLLIDE_TRACING_PASSWORD": "pass",
			},
			wants: Config{
				HTTP: HTTP{
					Port:     8088,
					GRPCPort: 8089,
				},
				Tracks: Tracks{
					Path: "tracks.yaml",
				},
				Logging: Logging{
					Level:      "DEBUG",
					WithSource: false,
					WithSpanID: false,
				},
				Tracing: Tracing{
					URI:      "localhost:8000",
					Username: "user",
					Password: "pass",
				},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			for key, value := range testcase.env {
				t.Setenv(key, value)
			}

			cfg, err := New()
			require.NoError(t, err)
			require.Equal(t, testcase.wants, cfg)
		})
	}
}
