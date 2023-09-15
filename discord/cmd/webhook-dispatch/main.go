package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/zalgonoise/x/pluslog"
	"github.com/zalgonoise/x/pluslog/httplog"

	"github.com/zalgonoise/x/discord/log"
	"github.com/zalgonoise/x/discord/webhook"
)

func main() {

}

func newLogger(webhookURL string) *slog.Logger {
	handlers := make([]slog.Handler, 0, 2)
	handlers = append(handlers,
		slog.NewTextHandler(os.Stderr, nil),
	)

	if webhookURL != "" {
		if _, _, err := webhook.Extract(webhookURL); err == nil {
			handlers = append(handlers, httplog.New(
				webhookURL, httplog.WithEncoder(log.New(log.JSON(true))),
			))
		}
	}

	return slog.New(pluslog.Multi(handlers...))
}

func run() (error, int) {
	configFile := flag.String("config", "", "path to the config.yaml file")
	modelsDir := flag.String("modelsDir", "", "path to the modelsDir directory")

	flag.Parse()

	config, err := newConfig(*configFile)
	if err != nil {
		return err, 1
	}

	logger := newLogger(config.LogsWebhookURL)
	ctx := context.Background()

	if *modelsDir != "" {
		config.ModelsPath = *modelsDir
	}

	models, err := newModels(config, logger)
	if err != nil {
		logger.ErrorContext(ctx, "loading models", "error", err.Error())

		return nil, 1
	}

	if err = models.Execute(ctx); err != nil {
		logger.ErrorContext(ctx, "executing webhook", "error", err.Error())

		return nil, 1
	}

	return nil, 0
}
