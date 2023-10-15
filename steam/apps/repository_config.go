package apps

import "github.com/zalgonoise/x/cfg"

type Config struct {
	uri string
}

func WithURI(uri string) cfg.Option[Config] {
	return cfg.Register[Config](func(config Config) Config {
		config.uri = uri

		return config
	})
}
