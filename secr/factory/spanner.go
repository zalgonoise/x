package factory

import (
	"io"
	"os"

	"github.com/zalgonoise/spanner"
)

const (
	traceFilePath = "/secr/trace.json"
)

// Spanner loads the file in the path `path`, to store the spanner entries in,
// defaulting to a std.err output if the path is empty or invalid
func Spanner(path string) {
	err := createTracefile(path)
	if err != nil {
		noTracefile()
	}
}

func createTracefile(p string) error {
	f, err := os.Create(p)
	if err != nil {
		if p == traceFilePath {
			return err
		}
		return createTracefile(traceFilePath)
	}

	spanner.To(spanner.Writer(io.MultiWriter(f, os.Stderr)))
	return nil
}

func noTracefile() {
	spanner.To(spanner.Writer(os.Stderr))
}
