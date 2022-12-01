package spanner

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/attr"
	"github.com/zalgonoise/logx/handlers/texth"
)

func runtime() {
	logger := logx.New(texth.New(os.Stderr))
	ctx := logx.InContext(context.Background(), logger)
	ctx, s := Start(ctx, "Runtime:Main")

	x := runtimeA(ctx, 2)
	runtimeE(ctx, "Hello", x)

	logx.From(ctx).Trace("Runtime:Main", s.End().AsAttr())
}

func runtimeA(ctx context.Context, i int) int {
	ctx, s := Start(ctx, "Runtime:A")

	x := i * 2

	logx.From(ctx).Trace("Runtime:A", s.End().AsAttr())
	return runtimeB(ctx, x)
}

func runtimeB(ctx context.Context, i int) int {
	ctx, s := Start(ctx, "Runtime:B")

	x := i * 2

	logx.From(ctx).Trace("Runtime:B", s.End().AsAttr())
	return runtimeC(ctx, x)
}

func runtimeC(ctx context.Context, i int) int {
	ctx, s := Start(ctx, "Runtime:C")

	x := i * 2

	logx.From(ctx).Trace("Runtime:C", s.End().AsAttr())

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

	logx.From(ctx).Trace("Runtime:E", s.End().AsAttr())
}
func TestFunctionsWithSpan(t *testing.T) {
	runtime()

	t.Error()
}
