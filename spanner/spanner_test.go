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

	_, s := Start(ctx, "Runtime:Start:A")
	x := runtimeA(ctx, 2)
	logx.From(ctx).Trace("Runtime:Start:A", AsAttr(s.End()))
	_, s = Start(ctx, "Runtime:Start:E")
	runtimeE(ctx, "Hello", x)
	logx.From(ctx).Trace("Runtime:Start:E", AsAttr(s.End()))

	logx.From(ctx).Trace("Runtime:Main", AsAttr(startS.End()))

	fmt.Print("\n\n\n")
	t := GetTrace(ctx)
	if t == nil {
		return
	}
	sp := t.Get()
	for _, s := range sp {
		fmt.Println(s.Extract())
	}
}

func runtimeA(ctx context.Context, i int) int {
	ctx, s := Start(ctx, "Runtime:A")

	s.Event("A: multiply by 2")
	x := i * 2

	logx.From(ctx).Trace("Runtime:A", AsAttr(s.End()))
	return runtimeB(ctx, x)
}

func runtimeB(ctx context.Context, i int) int {
	ctx, s := Start(ctx, "Runtime:B")

	s.Event("B: multiply by 2")
	x := i * 2

	logx.From(ctx).Trace("Runtime:B", AsAttr(s.End()))
	return runtimeC(ctx, x)
}

func runtimeC(ctx context.Context, i int) int {
	ctx, s := Start(ctx, "Runtime:C")

	s.Event("C: multiply by 2")
	x := i * 2

	logx.From(ctx).Trace("Runtime:C", AsAttr(s.End()))

	return x
}

func runtimeD(ctx context.Context, text string) {
	ctx, s := Start(ctx, "Runtime:D")

	fmt.Println(text)

	logx.From(ctx).Trace("Runtime:D", attr.New("span", s.End()))
}
func runtimeE(ctx context.Context, text string, i int) {
	ctx, s := Start(ctx, "Runtime:E")

	s.Add(attr.String("text", text), attr.Int("result", i))
	runtimeD(ctx, fmt.Sprintf("%s ; result: %v", text, i))

	logx.From(ctx).Trace("Runtime:E", AsAttr(s.End()))
}
func TestFunctionsWithSpan(t *testing.T) {
	runtime()

	t.Error()
}
