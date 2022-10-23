package main

import "github.com/zalgonoise/x/dns/cmd"

func main() {
	// reads OS env, CLI flags and config files
	// then runs the app based on that configuration
	//
	// blocking call; will error out if failed
	cmd.Run()
}
