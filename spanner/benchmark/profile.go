package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/x/spanner"
)

func main() {
	pprofFile, pprofErr := os.Create("/tmp/cpu.pprof")
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

	buf := new(bytes.Buffer)
	w := new(bytes.Buffer)
	spanner.To(spanner.Writer(buf))

	app := NewApp(w)
	for i := 0; i < 10_000; i++ {
		w.WriteString("3")

		ctx, span := spanner.Start(context.Background(), "pprof")

		err = app.Run(ctx)
		if err != nil {
			logx.Error("execution failed", attr.String("error", err.Error()))
			span.End()
			return
		}
		span.End()
		w.Reset()
	}
}

// Fibonacci returns the n-th fibonacci number.
func Fibonacci(n uint) (uint64, error) {
	if n <= 1 {
		return uint64(n), nil
	}

	var n2, n1 uint64 = 0, 1
	for i := uint(2); i < n; i++ {
		n2, n1 = n1, n1+n2
	}

	return n2 + n1, nil
}

// NewApp returns a new App.
func NewApp(r io.Reader) *App {
	return &App{r: r}
}

// App is a Fibonacci computation application.
type App struct {
	r io.Reader
}

// Run starts polling users for Fibonacci number requests and writes results.
func (a *App) Run(ctx context.Context) error {
	// Each execution of the run loop, we should get a new "root" span and context.
	newCtx, span := spanner.Start(ctx, "Run")

	n, err := a.Poll(newCtx)
	if err != nil {
		span.End()
		return err
	}

	a.Write(ctx, n)
	span.End()

	return nil

}

// Poll asks a user for input and returns the request.
func (a *App) Poll(ctx context.Context) (uint, error) {
	_, span := spanner.Start(ctx, "Poll")
	defer span.End()

	var n uint
	_, err := fmt.Fscanf(a.r, "%d\n", &n)

	// Store n as a string to not overflow an int64.
	span.Add(attr.Uint("input", n))

	return n, err
}

// Write writes the n-th Fibonacci number back to the user.
func (a *App) Write(ctx context.Context, n uint) {
	var span spanner.Span
	ctx, span = spanner.Start(ctx, "Write")
	defer span.End()

	f, err := func(ctx context.Context) (uint64, error) {
		_, span := spanner.Start(ctx, "Fibonacci")
		defer span.End()
		return Fibonacci(n)
	}(ctx)
	if err != nil {
		log.Printf("Fibonacci(%d): %v\n", n, err)
	} else {
		_ = f
		// log.Printf("Fibonacci(%d) = %d\n", n, f)
	}
}
