package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/zalgonoise/x/spanner"
)

const input string = "3"

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
}

func TestRuntime(t *testing.T) {
	buf := &bytes.Buffer{}
	w := &bytes.Buffer{}
	spanner.To(spanner.Writer(buf))

	//app
	app := NewApp(w)
	start := time.Now()

	//test
	for i := 0; i < 50_000; i++ {
		// Each execution of the run loop, we should get a new "root" span and context.
		ctx, span := spanner.Start(context.Background(), "Main")
		w.Reset()
		w.WriteString(input)
		err := app.Run(ctx)
		if err != nil {
			t.Errorf("execution failed: %v", err)
		}
		span.End()
		w.Reset()
	}

	t.Log(buf.Len())
	t.Log(buf.String())
	t.Error(time.Since(start))
}
