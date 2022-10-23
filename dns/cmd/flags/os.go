package flags

import (
	"os"

	"github.com/zalgonoise/x/dns/cmd/config"
)

// ParseOSEnv reads the contents of a number of target environment variables
// to create a new config from.
//
// Any unset variables will be left unset in the output config.Config
func ParseOSEnv() *config.Config {
	return &config.Config{
		DNS: &config.DNSConfig{
			Type:        os.Getenv("DNS_TYPE"),
			FallbackDNS: os.Getenv("DNS_FALLBACK"),
			Address:     os.Getenv("DNS_ADDRESS"),
			Prefix:      os.Getenv("DNS_PREFIX"),
			Proto:       os.Getenv("DNS_PROTO"),
		},
		Store: &config.StoreConfig{
			Type: os.Getenv("DNS_STORE_TYPE"),
			Path: os.Getenv("DNS_STORE_PATH"),
		},
		HTTP: &config.HTTPConfig{
			Port: intFromEnv("DNS_API_PORT"),
		},
		Logger: &config.LoggerConfig{
			Type: os.Getenv("DNS_LOGGER_TYPE"),
			Path: os.Getenv("DNS_LOGGER_PATH"),
		},
		Autostart: &config.AutostartConfig{
			DNS: boolFromEnv("DNS_AUTOSTART"),
		},
		Health: &config.HealthConfig{
			Type: os.Getenv("DNS_HEALTH_TYPE"),
		},
		Path: os.Getenv("DNS_CONFIG_PATH"),
	}
}
