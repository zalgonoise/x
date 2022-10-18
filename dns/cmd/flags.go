package cmd

import (
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"os"

	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/store"
	"gopkg.in/yaml.v2"
)

func ParseFlags() *config.Config {
	var conf = config.Default()
	var getFlags = true

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

	autostartDNS := flag.Bool("start-dns", true, "automatically start the DNS server")

	flag.Parse()

	if *configPath != "" {
		conf, getFlags = readConfig(*configPath, conf)
		defer writeConfig(conf, *configPath)
	}

	if getFlags {

		conf.Apply(
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

	return conf
}

func writeConfig(conf *config.Config, path string) {
	var (
		conv string
		b    []byte
		err  error
	)

	switch conf.Type {
	case "json":
		conv = conf.Type
		b, err = json.Marshal(conf)
	case "yaml":
		conv = conf.Type
		b, err = yaml.Marshal(conf)
	default:
		conf.Type = "yaml"
		b, err = yaml.Marshal(conf)
	}

	if err != nil {
		log.Printf("failed to encode config as %s: %v", conv, err)
		return
	}

	err = os.Remove(path)
	if err != nil {
		log.Printf("failed to remove old config file in %s: %v", path, err)
		return
	}

	err = os.WriteFile(path, b, fs.FileMode(store.OS_ALL_RWX))
	if err != nil {
		log.Printf("failed to write new config file in %s: %v", path, err)
		return
	}
}

func readConfig(path string, conf *config.Config) (*config.Config, bool) {
	var (
		ctype string
		jerr  error
		yerr  error
	)
	_, err := os.Stat(path)
	if err != nil {
		log.Printf("failed to stat config file in %s: %v", path, err)
		return conf, true
	}
	conf.Path = path

	b, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to read config file in %s: %v", path, err)
		return conf, true
	}
	if len(b) == 0 {
		return conf, true
	}

	jerr = json.Unmarshal(b, conf)
	switch jerr {
	case nil:
		ctype = "json"
	default:
		yerr = yaml.Unmarshal(b, conf)
	}

	if jerr != nil && yerr != nil {
		log.Printf(
			"failed to parse file content: JSON: %v ; YAML: %v", jerr, yerr,
		)
		return conf, true
	}

	if ctype == "" {
		ctype = "yaml"
	}

	conf.Type = ctype
	return conf, false
}
