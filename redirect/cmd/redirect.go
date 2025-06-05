package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/zalgonoise/x/redirect"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	readTimeout  = 5 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 120 * time.Second
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	exitCode, err := run(ctx, logger)
	if err != nil {
		logger.ErrorContext(ctx, "app runtime error", slog.String("error", err.Error()))
	}

	os.Exit(exitCode)
}

func run(ctx context.Context, logger *slog.Logger) (int, error) {
	cfg, err := redirect.NewConfig()
	if err != nil {
		return 1, err
	}

	handler := redirect.To(cfg.ToURI)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	mux := http.NewServeMux()
	mux.Handle("/", handler)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	slog.InfoContext(ctx, "set up HTTP server",
		slog.String("addr", addr),
		slog.String("redirect_to", cfg.ToURI),
		slog.Duration("read_timeout", readTimeout),
		slog.Duration("write_timeout", writeTimeout),
		slog.Duration("idle_timeout", idleTimeout),
	)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.ErrorContext(ctx, "listening and serving HTTP", slog.String("error", err.Error()))

			return
		}
	}()

	slog.InfoContext(ctx, "listening and serving HTTP")

	<-done
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	slog.InfoContext(ctx, "shutting down HTTP server")

	// Attempt to gracefully shut down the server.
	if err := server.Shutdown(ctx); err != nil {
		logger.ErrorContext(ctx, "shutting down HTTP server", slog.String("error", err.Error()))
	}

	return 0, nil
}
