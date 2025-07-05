package profiling

import (
	"context"
	"fmt"
	"log/slog"
)

func AsPyroscopeLogger(logger *slog.Logger) *PyroscopeLogger {
	return &PyroscopeLogger{logger}
}

type PyroscopeLogger struct {
	logger *slog.Logger
}

func (l *PyroscopeLogger) Infof(a string, b ...interface{}) {
	l.logger.InfoContext(context.Background(), fmt.Sprintf(a, b...))
}

func (l *PyroscopeLogger) Debugf(a string, b ...interface{}) {
	l.logger.DebugContext(context.Background(), fmt.Sprintf(a, b...))
}
func (l *PyroscopeLogger) Errorf(a string, b ...interface{}) {
	l.logger.ErrorContext(context.Background(), fmt.Sprintf(a, b...))
}
