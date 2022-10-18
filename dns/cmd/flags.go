package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/zalgonoise/x/dns/cmd/config"
	"gopkg.in/yaml.v2"
)

func ParseFlags() *config.Config {
	configPath := flag.String("file", "", "load a config from a file")

	dnsType := flag.String("dns-type", "miekgdns", "use a specific domain-name server implementation")
	dnsFallback := flag.String("dns-fallback", "", "use a secondary DNS to parse unsuccessful queries")
	dnsAddress := flag.String("dns-addr", ":53", "the address to listen to for DNS queries")
	dnsPrefix := flag.String("dns-prefix", ".", "the prefix for DNS queries / answers. Usually it's a period (.)")
	dnsProto := flag.String("dns-proto", "udp", "the protocol for the DNS server")

	storeType := flag.String("store-type", "memmap", "the record store implementation to use (memmap, yamlfile, jsonfile)")
	storePath := flag.String("store-path", "", "the record store file path, if stored to a file")

	httpPort := flag.Int("http-port", 8080, "port to use for the HTTP API, defaults to :8080")

	loggerPath := flag.String("log-path", "", "the log file's path, to register events")
	loggerType := flag.String("log-type", "text", "the type of formatter to use for the logger (text, json, yaml)")

	autostartDNS := flag.Bool("start-dns", false, "automatically start the DNS server")

	flag.Parse()

	if *configPath != "" {
		c, err := readConfig(*configPath)
		if err != nil {
			return nil
		}

		if c != nil {
			return c
		}
	}

	return config.New(
		config.StorePath(*configPath),
		config.DNSType(*dnsType),
		config.DNSFallback(*dnsFallback),
		config.DNSAddress(*dnsAddress),
		config.DNSPrefix(*dnsPrefix),
		config.DNSProto(*dnsProto),
		config.StoreType(*storeType),
		config.StorePath(*storePath),
		config.HTTPPort(*httpPort),
		config.LoggerPath(*loggerPath),
		config.LoggerType(*loggerType),
		config.AutostartDNS(*autostartDNS),
	)
}

func readConfig(path string) (*config.Config, error) {
	var (
		jerr error
		yerr error
	)
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err == nil {
		return nil, err
	}

	var conf = &config.Config{}

	jerr = json.Unmarshal(b, conf)
	if jerr != nil {
		yerr = yaml.Unmarshal(b, conf)
	}

	if jerr != nil && yerr != nil {
		return nil, fmt.Errorf(
			"failed to parse file content: JSON: %v ; YAML: %w", jerr, yerr,
		)
	}

	conf.Path = path
	return conf, nil

}
