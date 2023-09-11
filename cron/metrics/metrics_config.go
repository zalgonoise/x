package metrics

import "github.com/zalgonoise/x/cfg"

const (
	metricsViaProm = iota
)

type Config struct {
	metricsType int

	serverPort int
}

func ViaPrometheus() cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.metricsType = metricsViaProm

		return config
	})
}

func WithPort(port int) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.serverPort = port

		return config
	})
}
