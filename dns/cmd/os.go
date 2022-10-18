package cmd

import (
	"os"
	"strconv"

	"github.com/zalgonoise/x/dns/cmd/config"
)

func intFromEnv(s string) int {
	val := os.Getenv(s)
	if val == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func boolFromEnv(s string) bool {
	val := os.Getenv(s)
	return val != ""
}

func ParseOSEnv() *config.Config {
	return &config.Config{
		DNS: &config.DNSConfig{
			Type:    os.Getenv("DNS_TYPE"),
			Address: os.Getenv("DNS_ADDRESS"),
			Prefix:  os.Getenv("DNS_PREFIX"),
			Proto:   os.Getenv("DNS_PROTO"),
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
		Path: os.Getenv("DNS_CONFIG_PATH"),
	}
}
