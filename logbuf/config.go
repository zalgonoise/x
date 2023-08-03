package logbuf

import (
	"time"
)

// BufferedHandlerConfig describes the configuration values for a BufferedHandler
//
// TODO: extend and document data structure with config values as needed; add envars for fluidity
type BufferedHandlerConfig struct {
	DeletionThresh time.Duration
	FlushFrequency time.Duration
	FlushLevel     int
}
