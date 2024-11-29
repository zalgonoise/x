package main

import (
	"github.com/zalgonoise/x/cli"
	"github.com/zalgonoise/x/obs-midi/cmd/obs-midi-config/gen"
	"github.com/zalgonoise/x/obs-midi/cmd/obs-midi-config/validate"
)

var modes = []string{"validate", "gen"}

func main() {
	runner := cli.NewRunner("obs-midi-config",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"validate": cli.Executable(validate.Exec),
			"gen":      cli.Executable(gen.Exec),
		}),
	)

	cli.Run(runner)
}
