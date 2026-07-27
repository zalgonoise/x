package log

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		level string
		wants *slog.Logger
	}{
		{
			name:  "Valid/LowercaseDebug",
			level: "debug",
			wants: slog.New(&SpanContextHandler{
				withSpanID: true,
				handler: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
					AddSource: true,
					Level:     slog.LevelDebug,
				}),
			}),
		},
		{
			name:  "Valid/UppercaseError",
			level: "ERROR",
			wants: slog.New(&SpanContextHandler{
				withSpanID: true,
				handler: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
					AddSource: true,
					Level:     slog.LevelError,
				}),
			}),
		},
		{
			name:  "Invalid/UnknownLevelReturnsDefaults",
			level: "SERVER_ON_FIRE",
			wants: slog.New(&SpanContextHandler{
				withSpanID: true,
				handler: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
					AddSource: true,
					Level:     slog.LevelInfo,
				}),
			}),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.wants, New(testcase.level, true, true))
		})
	}
}
