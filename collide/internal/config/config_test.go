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
				Frontend: Frontend{
					Port: 8082,
				},
				Tracks: Tracks{
					Path: "",
				},
				Metrics: Metrics{
					URI: "collector:4318",
				},
				Logging: Logging{
					Level:      "INFO",
					WithSource: true,
					WithSpanID: true,
				},
				Tracing: Tracing{
					URI:      "tempo:4317",
					Username: "",
					Password: "",
				},
				Profiling: Profiling{
					Enabled: true,
					Name:    "collide",
					URI:     "http://pyroscope:4040",
					Tags: map[string]string{
						"hostname": "api.fallenpetals.com",
						"service":  "collide",
						"version":  "v1",
					},
				},
			},
		},
		{
			name: "Custom",
			env: map[string]string{
				"COLLIDE_HTTP_PORT":        "8088",
				"COLLIDE_GRPC_PORT":        "8089",
				"COLLIDE_FE_HTTP_PORT":     "8090",
				"COLLIDE_TRACKS_PATH":      "tracks.yaml",
				"COLLIDE_METRICS_URI":      "collector:14318",
				"COLLIDE_LOG_LEVEL":        "DEBUG",
				"COLLIDE_LOG_WITH_SOURCE":  "false",
				"COLLIDE_LOG_WITH_SPAN_ID": "false",
				"COLLIDE_TRACING_URI":      "localhost:8000",
				"COLLIDE_TRACING_USERNAME": "user",
				"COLLIDE_TRACING_PASSWORD": "pass",
				"COLLIDE_PROFILING_NAME":   "collide-dev",
				"COLLIDE_PROFILING_URI":    "http://pyroscope:14040",
				"COLLIDE_PROFILING_TAGS":   "hostname:api.dev.fallenpetals.com,service:collide,version:v2",
			},
			wants: Config{
				HTTP: HTTP{
					Port:     8088,
					GRPCPort: 8089,
				},
				Frontend: Frontend{
					Port: 8090,
				},
				Tracks: Tracks{
					Path: "tracks.yaml",
				},
				Metrics: Metrics{
					URI: "collector:14318",
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
				Profiling: Profiling{
					Enabled: true,
					Name:    "collide-dev",
					URI:     "http://pyroscope:14040",
					Tags: map[string]string{
						"hostname": "api.dev.fallenpetals.com",
						"service":  "collide",
						"version":  "v2",
					},
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
