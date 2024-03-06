package config

import "github.com/kelseyhightower/envconfig"

func Get() (*Config, error) {
	var config Config

	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
