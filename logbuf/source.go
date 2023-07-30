package logbuf

import (
	"runtime"

	"golang.org/x/exp/slog"
)

// Source returns a Source for the log event.
// If the Record was created without the necessary information,
// or if the location is unavailable, it returns a non-nil *Source
// with zero fields.
func Source(r slog.Record) slog.Source {
	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	return slog.Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}
