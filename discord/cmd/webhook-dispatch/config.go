package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/zalgonoise/x/discord/webhook"
)

const defaultConfigFile = "./config.yaml"

type Config struct {
	WebhookURL     string `yaml:"url,omitempty"`
	LogsWebhookURL string `yaml:"logs_url,omitempty"`
	Template       string `yaml:"template,omitempty"`
	ModelsPath     string `yaml:"models,omitempty"`
	Message        string `yaml:"message,omitempty"`
}

func validate(config *Config) error {
	if config.WebhookURL == "" {
		return errors.New("empty target webhook URL")
	}

	if _, _, err := webhook.Extract(config.WebhookURL); err != nil {
		return err
	}

	if config.LogsWebhookURL != "" {
		_, _, err := webhook.Extract(config.LogsWebhookURL)
		if err != nil {
			return err
		}
	}

	if config.Template == "" && config.Message == "" {
		return errors.New("must have at least a template selection or message content to execute in the webhook")
	}

	return nil
}

func newConfig(path string) (*Config, error) {
	if path == "" {
		path = defaultConfigFile
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file: %w", err)
	}

	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze configuration file: %w", err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, stat.Size()))

	_, err = buf.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	config := new(Config)
	if err = yaml.Unmarshal(buf.Bytes(), config); err != nil {
		return nil, fmt.Errorf("failed to extract configuration from file: %w", err)
	}

	if err := validate(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}
