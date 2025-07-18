package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTP      HTTP
	Frontend  Frontend
	Tracks    Tracks
	Metrics   Metrics
	Logging   Logging
	Tracing   Tracing
	Profiling Profiling
}

type HTTP struct {
	Port     int `env:"COLLIDE_HTTP_PORT" envDefault:"8080"`
	GRPCPort int `env:"COLLIDE_GRPC_PORT" envDefault:"8081"`
}

type Frontend struct {
	Port       int    `env:"COLLIDE_FE_HTTP_PORT" envDefault:"8082"`
	BackendURI string `env:"COLLIDE_BE_URI" envDefault:"http://api.fallenpetals.com:8083"`
}

type Tracks struct {
	Path string `env:"COLLIDE_TRACKS_PATH"`
}

type Metrics struct {
	URI string `env:"COLLIDE_METRICS_URI" envDefault:"collector:4318"`
}

type Logging struct {
	Level      string `env:"COLLIDE_LOG_LEVEL" envDefault:"INFO"`
	WithSource bool   `env:"COLLIDE_LOG_WITH_SOURCE" envDefault:"true"`
	WithSpanID bool   `env:"COLLIDE_LOG_WITH_SPAN_ID" envDefault:"true"`
}
type Tracing struct {
	URI      string `env:"COLLIDE_TRACING_URI" envDefault:"tempo:4317"`
	Username string `env:"COLLIDE_TRACING_USERNAME"`
	Password string `env:"COLLIDE_TRACING_PASSWORD"`
}

type Profiling struct {
	Enabled bool              `env:"COLLIDE_PROFILING_ENABLED" envDefault:"true"`
	Name    string            `env:"COLLIDE_PROFILING_NAME" envDefault:"collide"`
	URI     string            `env:"COLLIDE_PROFILING_URI" envDefault:"http://pyroscope:4040"`
	Tags    map[string]string `env:"COLLIDE_PROFILING_TAGS" envDefault:"hostname:api.fallenpetals.com,service:collide,version:v1"`
}

func New() (Config, error) {
	return env.ParseAs[Config]()
}
