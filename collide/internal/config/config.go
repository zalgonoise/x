package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Tracks  Tracks
	Logging Logging
	Tracing Tracing
}

type Tracks struct {
	Path string `env:"COLLIDE_TRACKS_PATH"`
}

type HTTP struct {
	Port     int `env:"COLLIDE_HTTP_PORT"`
	GRPCPort int `env:"COLLIDE_GRPC_PORT"`
}
type Logging struct {
	Level      string `env:"COLLIDE_LOG_LEVEL"`
	WithSource bool   `env:"COLLIDE_LOG_WITH_SOURCE"`
	WithSpanID bool   `env:"COLLIDE_LOG_WITH_SPAN_ID"`
}
type Tracing struct {
	URI      string `env:"COLLIDE_TRACING_URI"`
	Username string `env:"COLLIDE_TRACING_USERNAME"`
	Password string `env:"COLLIDE_TRACING_PASSWORD"`
}

func New() (*Config, error) {
	return env.ParseAs[*Config]()
}
