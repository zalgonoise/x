package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/switchupcb/dasgo/v10/dasgo"

	"github.com/zalgonoise/x/discord/webhook"
)

const (
	defaultModelsPath = "./models.json"
	defaultTimeout    = 15 * time.Second
)

func (m Models) Execute(ctx context.Context) error {
	for i := range m.ModelSet {
		if m.ModelSet[i].Name == m.config.Template {
			return m.exec(ctx, m.ModelSet[i].Content)
		}
	}

	if m.config.Message != "" {
		// execute webhook with message string
		return m.exec(ctx, []*dasgo.ExecuteWebhook{{
			Content: &m.config.Message,
		}})
	}

	return fmt.Errorf("template %s was not found in models and no message content was provided", m.config.Template)
}

func (m Models) exec(ctx context.Context, reqs []*dasgo.ExecuteWebhook) error {
	id, token, err := webhook.Extract(m.config.WebhookURL)
	if err != nil {
		return err
	}

	h, err := webhook.New(id, token, webhook.WithLogger(m.logger))
	if err != nil {
		return err
	}

	execCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	for idx := range reqs {
		res, execErr := h.Execute(execCtx, reqs[idx])

		if execErr != nil {
			if res != nil && res.Body != nil {
				res.Body.Close()
			}

			return execErr
		}

		res.Body.Close()
	}

	return nil
}

type Models struct {
	ModelSet []Model `json:"models"`

	config *Config
	logger *slog.Logger
}

type Model struct {
	Name    string                  `json:"name"`
	Content []*dasgo.ExecuteWebhook `json:"content"`
}

func newModels(config *Config, logger *slog.Logger) (*Models, error) {
	if config.ModelsPath == "" {
		config.ModelsPath = defaultModelsPath
	}

	f, err := os.Open(config.ModelsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open models file: %w", err)
	}

	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze models file: %w", err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, stat.Size()))

	_, err = buf.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read models file: %w", err)
	}

	models := new(Models)
	if err = json.Unmarshal(buf.Bytes(), models); err != nil {
		return nil, fmt.Errorf("failed to extract models from file: %w", err)
	}

	models.config = config
	models.logger = logger

	return models, nil
}
