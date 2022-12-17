package benchmark_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

// name is the Tracer name used to identify this instrumentation library.
const name = "fib"

const input string = "3"

func BenchmarkRuntime(b *testing.B) {
	w := new(bytes.Buffer)
	buf := new(bytes.Buffer)
	exp, err := newExporter(buf)
	if err != nil {
		b.Errorf("failed to create exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(newResource()),
	)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			b.Errorf("failed to shut down tracer: %v", err)
		}
	}()
	otel.SetTracerProvider(tp)

	//app
	app := NewApp(w)
	//test
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Each execution of the run loop, we should get a new "root" span and context.
		ctx, span := otel.Tracer(name).Start(context.Background(), "Main")
		w.WriteString(input)
		err := app.Run(ctx)
		if err != nil {
			b.Errorf("execution failed: %v", err)
		}
		w.Reset()
		span.End()
	}
	tp.ForceFlush(context.Background())
	b.StopTimer()
	b.Log(buf.String()[:256])
}

func TestRuntime(t *testing.T) {
	w := new(bytes.Buffer)
	b := new(bytes.Buffer)
	exp, err := newExporter(b)
	if err != nil {
		t.Errorf("failed to create exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(newResource()),
	)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			t.Errorf("failed to shut down tracer: %v", err)
		}
	}()
	otel.SetTracerProvider(tp)

	//app
	app := NewApp(w)
	//test
	// Each execution of the run loop, we should get a new "root" span and context.
	ctx, span := otel.Tracer(name).Start(context.Background(), "Main")
	w.Reset()
	w.WriteString(input)
	err = app.Run(ctx)
	if err != nil {
		t.Errorf("execution failed: %v", err)
	}
	span.End()
	w.Reset()

	tp.ForceFlush(context.Background())
	t.Log(b.String()[:256])

	t.Error()
}

// newExporter returns a console exporter.
func newExporter(w io.Writer) (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		stdouttrace.WithPrettyPrint(),
	)
}

// newResource returns a resource describing this application.
func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("fib"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	return r
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
	newCtx, span := otel.Tracer(name).Start(ctx, "Run")

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
	_, span := otel.Tracer(name).Start(ctx, "Poll")
	defer span.End()

	var n uint
	_, err := fmt.Fscanf(a.r, "%d\n", &n)

	// Store n as a string to not overflow an int64.
	nStr := strconv.FormatUint(uint64(n), 10)
	span.SetAttributes(attribute.String("request.n", nStr))

	return n, err
}

// Write writes the n-th Fibonacci number back to the user.
func (a *App) Write(ctx context.Context, n uint) {
	var span trace.Span
	ctx, span = otel.Tracer(name).Start(ctx, "Write")
	defer span.End()

	f, err := func(ctx context.Context) (uint64, error) {
		_, span := otel.Tracer(name).Start(ctx, "Fibonacci")
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
