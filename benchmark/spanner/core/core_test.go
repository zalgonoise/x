package core_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"testing"
)

// name is the Tracer name used to identify this instrumentation library.
const name = "fib"

var input = [8]string{
	"1", "2", "3", "4", "5", "6", "7", "8",
}

func BenchmarkRuntime(b *testing.B) {
	w := new(bytes.Buffer)

	//app
	app := NewApp(w)
	//test
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Reset()
		w.WriteString(input[2])
		err := app.Run(context.Background())
		if err != nil {
			b.Errorf("execution failed: %v", err)
		}
		w.Reset()
	}
}

func TestRuntime(t *testing.T) {
	w := new(bytes.Buffer)

	//app
	app := NewApp(w)
	//test
	for _, n := range input {
		w.Reset()
		w.WriteString(n)
		err := app.Run(context.Background())
		if err != nil {
			t.Errorf("execution failed: %v", err)
		}
		w.Reset()
	}
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
	n, err := a.Poll(ctx)
	if err != nil {
		return err
	}

	a.Write(ctx, n)
	return nil
}

// Poll asks a user for input and returns the request.
func (a *App) Poll(ctx context.Context) (uint, error) {
	var n uint
	_, err := fmt.Fscanf(a.r, "%d\n", &n)
	return n, err
}

// Write writes the n-th Fibonacci number back to the user.
func (a *App) Write(ctx context.Context, n uint) {
	f, err := Fibonacci(n)
	if err != nil {
		log.Printf("Fibonacci(%d): %v\n", n, err)
	} else {
		_ = f
		// log.Printf("Fibonacci(%d) = %d\n", n, f)
	}
}
