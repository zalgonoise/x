package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/zalgonoise/x/log"
	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/handlers/jsonh"
)

func main() {
	pprofFile, pprofErr := os.Create("cpu.pprof")
	if pprofErr != nil {
		fmt.Println(pprofErr)
		os.Exit(1)
	}
	err := pprof.StartCPUProfile(pprofFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer pprof.StopCPUProfile()

	buf := &bytes.Buffer{}
	l := log.New(jsonh.New(buf))
	a := []attr.Attr{
		attr.New("complex", true),
		attr.New("id", 1234567890),
		attr.New("content", []attr.Attr{attr.New("data", true)}),
		attr.New("affected", []string{"none", "nothing", "nada"}),
	}

	for i := 0; i < 5000; i++ {
		l.Info(
			"this is a long message describing a benchmark test log event",
			a...,
		)
	}
	fmt.Println(buf.String())
}
