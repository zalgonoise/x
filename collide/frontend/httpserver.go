package frontend

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"
)

//go:embed static/index.html
//go:embed static/collide_api.swagger.json
//go:embed static/swagger-ui.html
var staticFS embed.FS

type ShutdownFunc func(context.Context) error

func NewServer(ctx context.Context, backendURI string, port int, logger *slog.Logger) ShutdownFunc {
	mux := http.NewServeMux()
	mux.HandleFunc("/static/collide_api.swagger.json", serveFile("static/collide_api.swagger.json", "application/json", logger))
	mux.HandleFunc("/explore", serveFile("static/swagger-ui.html", "text/html; charset=utf-8", logger))
	mux.HandleFunc("/collide", NewCollisionsHandler(backendURI, logger))

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

func NewCollisionsHandler(uri string, logger *slog.Logger) http.HandlerFunc {
	// read the template content of the embedded file.
	htmlBytes, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		logger.ErrorContext(context.Background(), "reading template file", slog.String("err", err.Error()))

		os.Exit(1)
	}

	// prepare HTML template
	tmpl, err := template.New("collide-fe").Parse(string(htmlBytes))
	if err != nil {
		logger.ErrorContext(context.Background(), "parsing template", slog.String("err", err.Error()))

		os.Exit(1)
	}

	uriData := struct {
		BackendURI string
	}{
		BackendURI: uri,
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(htmlBytes)+len(uri)))

	// apply URI to template HTML
	if err := tmpl.Execute(buf, uriData); err != nil {
		logger.ErrorContext(context.Background(), "preparing HTML from template", slog.String("error", err.Error()))

		os.Exit(1)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// set the correct Content-Type header so the browser renders it as HTML.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(buf.Bytes()); err != nil {
			logger.ErrorContext(r.Context(), "writing bytes to response", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error: failed to write response.", http.StatusInternalServerError)
		}
	}
}

func serveFile(filePath, contentType string, logger *slog.Logger) http.HandlerFunc {
	buf, err := staticFS.ReadFile(filePath)
	if err != nil {
		logger.ErrorContext(context.Background(), "reading embedded file", slog.String("err", err.Error()))

		os.Exit(1)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write(buf)
	}
}
