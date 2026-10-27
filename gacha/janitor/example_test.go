package janitor_test

import (
	"context"
	"fmt"
	"time"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/janitor"
	"github.com/zalgonoise/x/gacha/memory"
)

func Example() {
	ctx := context.Background()
	repo := memory.New[string]()

	// Move one entry to the dead-letter table.
	if err := repo.Write(ctx, "poison"); err != nil {
		fmt.Println(err)

		return
	}

	entries, err := repo.Claim(ctx, 1, time.Minute)
	if err != nil {
		fmt.Println(err)

		return
	}

	err = repo.Settle(ctx, gacha.Outcome{
		ID:       entries[0].ID,
		Attempts: entries[0].Attempts,
		Status:   gacha.StatusFailed,
		Err:      "cannot be processed",
	})
	if err != nil {
		fmt.Println(err)

		return
	}

	// Archive dead letters (here, print them), then delete them. Run would repeat this every interval.
	archive := func(_ context.Context, data string) error {
		fmt.Println("archiving", data)

		return nil
	}

	res, err := janitor.New(repo.DeadLetters(0), archive, janitor.WithInterval(time.Hour)).Poll(ctx)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println("purged", res.Completed)

	// Output:
	// archiving poison
	// purged 1
}
