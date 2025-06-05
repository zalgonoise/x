package redirect

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Host  string `env:"REDIRECT_HOST" envDefault:"localhost"`
	Port  int    `env:"REDIRECT_PORT" envDefault:"8080"`
	ToURI string `env:"REDIRECT_URI" envDefault:"google.com"`
}

func NewConfig() (Config, error) {
	return env.ParseAs[Config]()
}
