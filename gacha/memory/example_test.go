package memory_test

import (
	"context"
	"fmt"
	"time"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/memory"
)

func Example() {
	ctx := context.Background()

	// With gacha.Identity, the repository stores the items themselves, without marshaling them.
	repo := memory.New[string]()

	if err := repo.Write(ctx, "first", "second"); err != nil {
		fmt.Println(err)

		return
	}

	entries, err := repo.Claim(ctx, 10, time.Minute)
	if err != nil {
		fmt.Println(err)

		return
	}

	outcomes := make([]gacha.Outcome, 0, len(entries))

	for _, entry := range entries {
		fmt.Println(entry.Data, entry.Attempts)

		outcomes = append(outcomes, gacha.Outcome{
			ID:       entry.ID,
			Attempts: entry.Attempts,
			Status:   gacha.StatusCompleted,
		})
	}

	if err = repo.Settle(ctx, outcomes...); err != nil {
		fmt.Println(err)

		return
	}

	stats, err := repo.Stats(ctx)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Printf("%+v\n", stats)

	// Output:
	// first 1
	// second 1
	// {Pending:0 Processing:0 Completed:2 DeadLetters:0}
}
