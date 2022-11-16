package benchmark

import (
	"bytes"
	"testing"

	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/handlers/jsonh"
	"github.com/zalgonoise/x/log/handlers/texth"

	logv2 "github.com/zalgonoise/x/log"
)

func BenchmarkLogger(b *testing.B) {
	const (
		prefix  = "benchmark"
		sub     = "test"
		msg     = "benchmark test log event"
		longMsg = "this is a long message describing a benchmark test log event"
	)

	var (
		newMeta = []attr.Attr{
			attr.NewAttr("complex", true),
			attr.NewAttr("id", 1234567890),
			attr.NewAttr("content", []attr.Attr{attr.NewAttr("data", true)}),
			attr.NewAttr("affected", []string{"none", "nothing", "nada"}),
		}
		buf = new(bytes.Buffer)
	)

	b.Run("Writing", func(b *testing.B) {
		b.Run("SimpleText", func(b *testing.B) {
			b.Run("LogV2", func(b *testing.B) {
				localLogger := logv2.New(texth.New(buf))

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info(msg)
				}
				buf.Reset()
			})
		})
		b.Run("SimpleJSON", func(b *testing.B) {
			b.Run("LogV2", func(b *testing.B) {
				localLogger := logv2.New(jsonh.New(buf))

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info(msg)
				}
				buf.Reset()
			})
		})

		b.Run("ComplexText", func(b *testing.B) {
			b.Run("LogV2", func(b *testing.B) {
				localLogger := logv2.New(texth.New(buf))

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info(longMsg, newMeta...)
				}
				buf.Reset()
			})
		})
		b.Run("ComplexJSON", func(b *testing.B) {
			b.Run("LogV2", func(b *testing.B) {
				localLogger := logv2.New(jsonh.New(buf))

				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					localLogger.Info(longMsg, newMeta...)
				}
				buf.Reset()
			})
		})
	})
}
