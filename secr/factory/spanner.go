package factory

import (
	"io"
	"os"

	"github.com/zalgonoise/spanner"
)

const (
	traceFilePath = "/secr/trace.json"
)

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
