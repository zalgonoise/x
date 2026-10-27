package writer_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/memory"
	"github.com/zalgonoise/x/gacha/writer"
)

type Order struct {
	ID    int    `json:"id"`
	Total string `json:"total"`
}

func ExampleNewSync() {
	ctx := context.Background()
	repo := memory.New[[]byte]()

	w := writer.NewSync(repo, gacha.JSON[Order]())

	if err := w.Write(ctx, Order{ID: 1, Total: "9.99"}, Order{ID: 2, Total: "0.50"}); err != nil {
		fmt.Println(err)

		return
	}

	entries, err := repo.Claim(ctx, 10, time.Minute)
	if err != nil {
		fmt.Println(err)

		return
	}

	for _, entry := range entries {
		fmt.Println(string(entry.Data))
	}

	// Output:
	// {"id":1,"total":"9.99"}
	// {"id":2,"total":"0.50"}
}

func ExampleNewBatch() {
	ctx := context.Background()
	repo := memory.New[[]byte]()

	// Concurrent writes within 5ms of each other share a single repository write.
	w := writer.NewBatch(repo, gacha.JSON[Order](), writer.WithLinger(5*time.Millisecond))

	var wg sync.WaitGroup

	for id := range 10 {
		wg.Go(func() {
			if err := w.Write(ctx, Order{ID: id}); err != nil {
				fmt.Println(err)
			}
		})
	}

	wg.Wait()

	if err := w.Close(ctx); err != nil {
		fmt.Println(err)

		return
	}

	stats, err := repo.Stats(ctx)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(stats.Pending)

	// Output:
	// 10
}
