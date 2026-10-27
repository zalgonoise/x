//go:build integration

package postgres_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/gachatest"
	"github.com/zalgonoise/x/gacha/postgres"
)

func benchData(n int) [][]byte {
	data := make([][]byte, n)

	for i := range data {
		data[i] = gachatest.Bytes(i)
	}

	return data
}

func completed(entries []gacha.Entry[[]byte]) []gacha.Outcome {
	outcomes := make([]gacha.Outcome, len(entries))

	for i := range entries {
		outcomes[i] = gacha.Outcome{ID: entries[i].ID, Attempts: entries[i].Attempts, Status: gacha.StatusCompleted}
	}

	return outcomes
}

func BenchmarkRepository_Write(b *testing.B) {
	for _, size := range []int{1, 100, 1000} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			repo := newRepo(b)
			data := benchData(size)

			b.ReportAllocs()

			for b.Loop() {
				if err := repo.Write(b.Context(), data...); err != nil {
					b.Fatal(err)
				}
			}

			b.ReportMetric(float64(b.N*size)/b.Elapsed().Seconds(), "entries/s")
		})
	}
}

// BenchmarkRepository_Cycle writes, claims and settles a batch per iteration, settling it in one call (Batch) or one
// call per entry (PerItem), the poller's default.
func BenchmarkRepository_Cycle(b *testing.B) {
	for _, size := range []int{1, 100} {
		for _, mode := range []string{"Batch", "PerItem"} {
			b.Run(strconv.Itoa(size)+"/"+mode, func(b *testing.B) {
				repo := newRepo(b, postgres.WithDeleteOnComplete())
				data := benchData(size)

				b.ReportAllocs()

				for b.Loop() {
					if err := repo.Write(b.Context(), data...); err != nil {
						b.Fatal(err)
					}

					entries, err := repo.Claim(b.Context(), size, time.Minute)
					if err != nil {
						b.Fatal(err)
					}

					outcomes := completed(entries)

					if mode == "Batch" {
						if err = repo.Settle(b.Context(), outcomes...); err != nil {
							b.Fatal(err)
						}

						continue
					}

					for i := range outcomes {
						if err = repo.Settle(b.Context(), outcomes[i]); err != nil {
							b.Fatal(err)
						}
					}
				}

				b.ReportMetric(float64(b.N*size)/b.Elapsed().Seconds(), "entries/s")
			})
		}
	}
}
