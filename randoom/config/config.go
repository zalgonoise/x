package config

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/zalgonoise/x/randoom/items"
)

const (
	configJSONFilename = "config.json"
	configYAMLFilename = "config.yaml"
)

var ErrInvalidConfig = errors.New("invalid config")

type Config struct {
	DatabaseURI string     `json:"db_uri" yaml:"db_uri"`
	PlaylistID  string     `json:"playlist_id" yaml:"playlist_id"`
	Content     items.List `json:"content" yaml:"content"`
}

func OpenConfigFile(dir string) (cfg *Config, err error) {
	// use wd if empty
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	// add trailing slash
	if dir != "" && dir[len(dir)-1] != '\\' {
		dir = dir + `\`
	}

	cfg = &Config{}

	// read json
	f, err := os.Open(dir + configJSONFilename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if err == nil {
		defer f.Close()

		buf, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(buf, cfg); err == nil {
			return cfg, nil
		}
	}

	// read yaml
	f, err = os.Open(dir + configYAMLFilename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if err == nil {
		defer f.Close()

		buf, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		if err = yaml.Unmarshal(buf, cfg); err == nil {
			return cfg, nil
		}
	}

	return nil, ErrInvalidConfig
}

func ParseContent(dbURI, data, id string) (*Config, error) {
	list := &items.List{}

	if err := json.Unmarshal([]byte(data), list); err == nil {
		return &Config{
			DatabaseURI: dbURI,
			Content:     *list,
			PlaylistID:  id,
		}, nil
	}

	if err := yaml.Unmarshal([]byte(data), list); err == nil {
		return &Config{
			DatabaseURI: dbURI,
			Content:     *list,
			PlaylistID:  id,
		}, nil
	}

	return nil, ErrInvalidConfig
}
