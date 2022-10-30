package flags

import (
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/store/encoder"
)

// writeConfig writes config.Config `conf` to file in path `path`, according to `conf`'s Type
func writeConfig(conf *config.Config, path string) {
	var (
		conv string
		b    []byte
		err  error
	)

	if conf.Type == "" {
		conf.Type = "yaml"
	}
	enc := encoder.New(conf.Type)
	b, err = enc.Encode(conf)

	if err != nil {
		log.Printf("failed to encode config as %s: %v", conv, err)
		return
	}

	err = os.WriteFile(path, b, fs.FileMode(store.OS_ALL_RW))
	if err != nil {
		log.Printf("failed to write new config file in %s: %v", path, err)
		return
	}
}

// readConfig will read the input config.Config in path `path`, and unmarshal it into
// config.Config `conf`.
//
// It returns the resulting config.Config and an `ok` bool. The `ok` bool serves as a trigger
// for getting CLI flags or not.
//
// If the config file does not exist, it is created (returning the initial config and `true`
// if the operations fail)
//
// if the config file exists and is invalid (can't unmarshal as either JSON or YAML),
// it will be overwritten; returning the initial config and `true`
//
// if the config file exists and is valid, returns the parsed config and `false`
func readConfig(path string, conf *config.Config) (*config.Config, bool) {
	var ()
	_, err := os.Stat(path)
	if err != nil {
		f, err := os.Create(path)
		if err != nil {
			log.Printf("failed to stat config file in %s: %v", path, err)
			return conf, true
		}
		err = f.Sync()
		if err != nil {
			log.Printf("failed to save file to disk in %s: %v", path, err)
			return conf, true
		}
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

	etype, err := decodeBuffer(b, conf)
	if err != nil {
		log.Print(err)
	}
	conf.Type = etype

	return conf, false
}

func decodeBuffer(b []byte, rec interface{}) (etype string, err error) {
	var (
		jerr error
		yerr error
		jenc = encoder.New("json")
		yenc = encoder.New("yaml")
	)

	jerr = jenc.Decode(b, rec)
	if jerr != nil {
		yerr = yenc.Decode(b, rec)
		if yerr != nil {
			err = fmt.Errorf(
				"failed to parse file content: JSON: %v ; YAML: %v", jerr, yerr,
			)
			return
		} else {
			etype = "yaml"
		}
	} else {
		etype = "json"
	}

	return
}
