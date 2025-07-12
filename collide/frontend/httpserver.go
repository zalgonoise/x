package frontend

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// The go:embed directive tells the Go compiler to embed the index.html file
// into the 'staticFS' variable. This creates a self-contained binary.
// It's required that the file exists in the same directory or a subdirectory.
//
//go:embed index.html
var staticFS embed.FS

type ShutdownFunc func(context.Context) error

func NewServer(ctx context.Context, port int, logger *slog.Logger) ShutdownFunc {
	mux := http.NewServeMux()
	mux.Handle("/collide", NewCollisionsHandler(logger))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    1024,
	}

	slog.InfoContext(ctx, "starting HTTP server", slog.Int("port", port))

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.ErrorContext(ctx, "running http server", slog.String("err", err.Error()))

			os.Exit(1)
		}
	}()

	return server.Shutdown
}

func NewCollisionsHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the content of the embedded file.
		htmlBytes, err := staticFS.ReadFile("index.html")
		if err != nil {
			// This should not happen if the file is embedded correctly at compile time.
			logger.ErrorContext(context.Background(), "serving HTML file", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error: could not serve page.", http.StatusInternalServerError)

			return
		}

		// Set the correct Content-Type header so the browser renders it as HTML.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Write the file content to the response body.
		if _, err := w.Write(htmlBytes); err != nil {
			logger.ErrorContext(context.Background(), "writing bytes to response", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error: failed to write response.", http.StatusInternalServerError)
		}
	}
}
