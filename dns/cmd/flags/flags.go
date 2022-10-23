package flags

import (
	"flag"

	"github.com/zalgonoise/x/dns/cmd/config"
)

// ParseFlags will read the input OS environment variables, CLI flags, and config file
// to generate and return a config.Config
func ParseFlags() *config.Config {
	var conf = config.Default()
	var getFlags = true
	var confPath string

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

	healthType := flag.String("health-type", "simplehealth", "the type of health / status report (simplehealth)")

	autostartDNS := flag.Bool("start-dns", true, "automatically start the DNS server")

	flag.Parse()
	osFlags := ParseOSEnv()

	if *configPath != "" {
		confPath = *configPath
	}
	if osFlags.Path != "" {
		confPath = osFlags.Path // OS vars overwrite CLI flags
	}

	if confPath != "" {
		conf, getFlags = readConfig(confPath, conf)
		defer writeConfig(conf, confPath)
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
			config.HealthType(*healthType),
			config.AutostartDNS(*autostartDNS),
		)
	}

	return config.Merge(conf, osFlags)
}
