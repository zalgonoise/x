package profiling

import (
	"log/slog"

	"github.com/grafana/pyroscope-go"
)

func New(name, uri string, tags map[string]string, logger *slog.Logger) (*pyroscope.Profiler, error) {
	return pyroscope.Start(pyroscope.Config{
		ApplicationName: name,
		ServerAddress:   uri,
		Logger:          AsPyroscopeLogger(logger),
		Tags:            tags,
		ProfileTypes: []pyroscope.ProfileType{
			// these profile types are enabled by default:
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			// these profile types are optional:
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
}
