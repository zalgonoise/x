package spanner

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/attr"
	"github.com/zalgonoise/logx/handlers/jsonh"
)

func runtime() {
	logger := logx.New(jsonh.New(os.Stderr))
	ctx := logx.InContext(context.Background(), logger)
	ctx, startS := Start(ctx, "Runtime:Main")

	// _, s := Start(ctx, "Runtime:Start:A")
	x := runtimeA(ctx, 2)
	// logx.From(ctx).Trace("Runtime:Start:A", AsAttr(s.End()))
	// _, s = Start(ctx, "Runtime:Start:E")
	runtimeE(ctx, "Hello", x)
	// logx.From(ctx).Trace("Runtime:Start:E", AsAttr(s.End()))

	startS.End()
	logx.From(ctx).Trace("Runtime:Main", AsAttr(startS.Extract()))

	fmt.Print("\n\n\n")
	t := GetTrace(ctx)
	if t == nil {
		return
	}
	sp := t.Extract()
	for _, s := range sp {
		fmt.Println(s)
	}
}

func runtimeA(ctx context.Context, i int) int {
	ctx, s := Start(ctx, "Runtime:A")
	defer logx.From(ctx).Trace("Runtime:A", AsAttr(s.Extract()))
	defer s.End()

	s.Event("A: multiply by 2")
	x := i * 2

	return runtimeB(ctx, x)
}

func runtimeB(ctx context.Context, i int) int {
	ctx, s := Start(ctx, "Runtime:B")
	defer logx.From(ctx).Trace("Runtime:B", AsAttr(s.Extract()))
	defer s.End()

	s.Event("B: multiply by 2")
	x := i * 2

	return runtimeC(ctx, x)
}

func runtimeC(ctx context.Context, i int) int {
	ctx, s := Start(ctx, "Runtime:C")
	defer s.End()

	s.Event("C: multiply by 2")
	x := i * 2

	logx.From(ctx).Trace("Runtime:C", AsAttr(s.Extract()))

	return x
}

func runtimeD(ctx context.Context, text string) {
	ctx, s := Start(ctx, "Runtime:D")
	defer s.End()

	fmt.Println(text)

	logx.From(ctx).Trace("Runtime:D", attr.New("span", s.Extract()))
}
func runtimeE(ctx context.Context, text string, i int) {
	ctx, s := Start(ctx, "Runtime:E")
	defer s.End()

	s.Add(attr.String("text", text), attr.Int("result", i))
	runtimeD(ctx, fmt.Sprintf("%s ; result: %v", text, i))

	logx.From(ctx).Trace("Runtime:E", AsAttr(s.Extract()))
}
func TestFunctionsWithSpan(t *testing.T) {
	runtime()

	t.Error()
}

func TestMainSpan(t *testing.T) {
	ctx := logx.InContext(context.Background(), logx.Default())
	ctx, s := Start(ctx, "main")
	defer func() {
		s.End()
		logx.From(ctx).Trace("trace", attr.New("spans", Extract(ctx)))
	}()

	t.Error()
}
