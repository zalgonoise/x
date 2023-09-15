package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

func main() {
	err, code := run()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to execute webhook-dispatch: %s", err.Error())
	}

	os.Exit(code)
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

	logger.InfoContext(ctx, "loading models")

	models, err := newModels(config, logger)
	if err != nil {
		logger.ErrorContext(ctx, "loading models", "error", err.Error())

		return nil, 1
	}

	logger.InfoContext(ctx, "executing webhook")

	if err = models.Execute(ctx); err != nil {
		logger.ErrorContext(ctx, "executing webhook", "error", err.Error())

		return nil, 1
	}

	return nil, 0
}
