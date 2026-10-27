package memory_test

import (
	"strconv"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/gachatest"
	"github.com/zalgonoise/x/gacha/memory"
)

// Compile-time checks that Repository implements every gacha capability.
var (
	_ gacha.Repository[[]byte] = (*memory.Repository[[]byte])(nil)
	_ gacha.Extender           = (*memory.Repository[[]byte])(nil)
	_ gacha.Redriver           = (*memory.Repository[[]byte])(nil)
	_ gacha.Stater             = (*memory.Repository[[]byte])(nil)
	_ gacha.Viewer[[]byte]     = (*memory.Repository[[]byte])(nil)
)

type order struct {
	ID    int
	Total string
}

func TestRepository(t *testing.T) {
	gachatest.Suite[[]byte]{
		New: func(*testing.T) gacha.Repository[[]byte] {
			return memory.New[[]byte]()
		},
		NewDeleteOnComplete: func(*testing.T) gacha.Repository[[]byte] {
			return memory.New[[]byte](memory.WithDeleteOnComplete())
		},
		Value: gachatest.Bytes,
		Wrap:  synctest.Test,
	}.Run(t)
}

func TestRepository_RealClock(t *testing.T) {
	if testing.Short() {
		t.Skip("waits for real leases")
	}

	gachatest.Suite[[]byte]{
		New: func(*testing.T) gacha.Repository[[]byte] {
			return memory.New[[]byte]()
		},
		Value: gachatest.Bytes,
		Lease: 20 * time.Millisecond,
	}.Run(t)
}

func TestRepository_Struct(t *testing.T) {
	gachatest.Suite[order]{
		New: func(*testing.T) gacha.Repository[order] {
			return memory.New[order]()
		},
		Value: func(i int) order {
			return order{ID: i, Total: strconv.Itoa(i * 100)}
		},
		Wrap: synctest.Test,
	}.Run(t)
}

func TestRepository_StoresValuesAsGiven(t *testing.T) {
	repo := memory.New[[]byte]()
	value := []byte("value")

	require.NoError(t, repo.Write(t.Context(), value))

	entries, err := repo.Claim(t.Context(), 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Same(t, &value[0], &entries[0].Data[0])
}

func BenchmarkRepository(b *testing.B) {
	for _, size := range []int{1, 100, 1000} {
		data := make([][]byte, size)

		for i := range data {
			data[i] = gachatest.Bytes(i)
		}

		// Each iteration writes, claims and settles a batch, so the repository stays at a steady size.
		b.Run("Cycle/"+strconv.Itoa(size), func(b *testing.B) {
			repo := memory.New[[]byte](memory.WithDeleteOnComplete())
			outcomes := make([]gacha.Outcome, 0, size)

			b.ReportAllocs()

			for b.Loop() {
				if err := repo.Write(b.Context(), data...); err != nil {
					b.Fatal(err)
				}

				entries, err := repo.Claim(b.Context(), size, time.Minute)
				if err != nil {
					b.Fatal(err)
				}

				outcomes = outcomes[:0]
				for i := range entries {
					outcomes = append(outcomes, gacha.Outcome{
						ID:       entries[i].ID,
						Attempts: entries[i].Attempts,
						Status:   gacha.StatusCompleted,
					})
				}

				if err = repo.Settle(b.Context(), outcomes...); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
