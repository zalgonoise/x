package mapping_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/zalgonoise/x/mapping"
)

func generateBenchmarkIntervals(n int) []mapping.DataInterval[int] {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	items := make([]mapping.DataInterval[int], n)

	for i := range n {
		// Create staggered overlapping intervals.
		startOffset := time.Duration((i*37)%n) * time.Hour
		duration := time.Duration(3+(i%7)) * time.Hour

		items[i] = mapping.DataInterval[int]{
			Interval: mapping.Interval{
				From: base.Add(startOffset),
				To:   base.Add(startOffset + duration),
			},
			Data: i % 5,
		}
	}

	return items
}

func BenchmarkOrganize(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		data := generateBenchmarkIntervals(size)

		b.Run(fmt.Sprintf("Replace/N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = mapping.Organize(data, mapping.WithReplace[int]())
			}
		})

		b.Run(fmt.Sprintf("Flatten/N=%d", size), func(b *testing.B) {
			mergeFn := func(cur, next int) int { return cur + next }
			b.ReportAllocs()
			for b.Loop() {
				_ = mapping.Organize(data, mapping.WithFlatten(mergeFn))
			}
		})

		b.Run(fmt.Sprintf("Keep/N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = mapping.Organize(data, mapping.WithKeep[int]())
			}
		})

		b.Run(fmt.Sprintf("Append/N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = mapping.Organize(data, mapping.WithAppend[int]())
			}
		})
	}
}

func BenchmarkTimeframeStream(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		data := generateBenchmarkIntervals(size)
		mergeFn := func(cur, next int) int { return cur + next }

		b.Run(fmt.Sprintf("Flatten/N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				stream := mapping.NewTimeframeStream(mapping.WithFlatten(mergeFn))
				for i := range data {
					stream.Add(data[i].Interval, data[i].Data)
				}
			}
		})
	}
}
