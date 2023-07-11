package main

import (
	"os"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/handlers/texth"

	"github.com/zalgonoise/x/audio/cmd/audio/config"
)

func main() {
	err, code := run()
	if err != nil {
		logx.New(texth.New(os.Stderr)).Error(
			"audio/gsp: runtime error",
			attr.String("error", err.Error()),
		)
	}
	os.Exit(code)
}

func run() (error, int) {
	_, err := config.NewConfig()
	if err != nil {
		return err, 1
	}

	return nil, 0
}
