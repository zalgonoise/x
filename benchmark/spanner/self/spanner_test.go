package benchmark_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/x/spanner"
)

var input string = "3"

func BenchmarkRuntime(b *testing.B) {
	buf := new(bytes.Buffer)
	w := new(bytes.Buffer)
	spanner.To(spanner.Writer(buf))

	//app
	app := NewApp(w)
	//test
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, span := spanner.Start(context.Background(), "Main")

		// Each execution of the run loop, we should get a new "root" span and context.
		w.WriteString(input)
		err := app.Run(ctx)
		if err != nil {
			b.Errorf("execution failed: %v", err)
		}
		w.Reset()
		span.End()
	}
	b.StopTimer()
}

func TestRuntime(t *testing.T) {
	b := new(bytes.Buffer)
	w := new(bytes.Buffer)
	spanner.To(spanner.Writer(b))

	//app
	app := NewApp(w)
	//test
	// Each execution of the run loop, we should get a new "root" span and context.
	ctx, span := spanner.Start(context.Background(), "Main")
	w.WriteString(input)
	err := app.Run(ctx)
	if err != nil {
		t.Errorf("execution failed: %v", err)
	}
	span.End()
	w.Reset()

	spanner.Processor().Flush(context.Background())
	time.Sleep(10 * time.Millisecond) // safety sleep after flushing
	t.Log(b.String()[:256])

	t.Error()
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

// App is a Fibonacci computation application.
type App struct {
	r io.Reader
}

// NewApp returns a new App.
func NewApp(r io.Reader) *App {
	return &App{r: r}
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
