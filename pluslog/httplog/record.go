package httplog

import (
	"log/slog"
	"time"
)

type HTTPRecord struct {
	Time    time.Time      `json:"timestamp"`
	Message string         `json:"message"`
	Level   string         `json:"level"`
	Source  *slog.Source   `json:"source,omitempty"`
	Attrs   map[string]any `json:"attributes,omitempty"`
}
