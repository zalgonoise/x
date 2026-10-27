package writer_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/writer"
)

// BenchmarkWriters compares Sync and Batch under concurrent writes, against a sink without latency and one that
// simulates database commits: calls are serialized and each one takes a fixed time, like a WAL flush.
func BenchmarkWriters(b *testing.B) {
	for _, latency := range []time.Duration{0, 100 * time.Microsecond} {
		var mu sync.Mutex

		sink := gacha.SinkFunc[string](func(context.Context, ...string) error {
			if latency > 0 {
				mu.Lock()
				time.Sleep(latency)
				mu.Unlock()
			}

			return nil
		})

		prefix := "Latency/" + latency.String()

		b.Run(prefix+"/Sync", func(b *testing.B) {
			w := writer.NewSync(sink, itoa)

			benchmarkWriter(b, w)
		})

		b.Run(prefix+"/Batch", func(b *testing.B) {
			w := writer.NewBatch(sink, itoa)
			defer func() {
				if err := w.Close(context.Background()); err != nil {
					b.Fatal(err)
				}
			}()

			benchmarkWriter(b, w)
		})
	}
}

func benchmarkWriter(b *testing.B, w gacha.Writer[int]) {
	b.Helper()
	b.ReportAllocs()
	b.SetParallelism(8)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := w.Write(context.Background(), 1, 2, 3); err != nil {
				b.Error(err)
			}
		}
	})
}
