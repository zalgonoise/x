package factory

import (
	"os"

	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/handlers"
	"github.com/zalgonoise/logx/handlers/jsonh"
	"github.com/zalgonoise/logx/handlers/texth"
)

const (
	logFilePath = "/secr/error.log"
)

func Logger(path string) logx.Logger {
	l, err := createLogfile(path)
	if err != nil {
		return noLogfile()
	}
	return l
}

func createLogfile(p string) (logx.Logger, error) {
	f, err := os.Create(p)
	if err != nil {
		if p == logFilePath {
			return nil, err
		}
		return createLogfile(logFilePath)
	}
	return logx.New(handlers.Multi(
		texth.New(os.Stdout),
		jsonh.New(f),
	)), nil
}

func noLogfile() logx.Logger {
	return logx.New(texth.New(os.Stdout))
}
