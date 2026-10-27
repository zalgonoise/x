package poller_test

import (
	"context"
	"fmt"
	"time"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/memory"
	"github.com/zalgonoise/x/gacha/poller"
	"github.com/zalgonoise/x/gacha/writer"
)

func Example() {
	type Order struct {
		ID int `json:"id"`
	}

	ctx := context.Background()
	repo := memory.New[[]byte]()

	if err := writer.NewSync(repo, gacha.JSON[Order]()).Write(ctx, Order{ID: 1}, Order{ID: 2}); err != nil {
		fmt.Println(err)

		return
	}

	// T = Order and R = []byte are inferred from the repository and the codec.
	p := poller.New(repo, gacha.JSON[Order](), func(_ context.Context, o Order) error {
		fmt.Println("handling order", o.ID)

		return nil
	},
		poller.WithBatchSize(100),
		poller.WithLease(time.Minute),
		poller.WithMaxAttempts(5),
		poller.WithRetryBackoff(poller.Exponential(time.Second, time.Hour)),
	)

	// Poll runs a single claim → handle → settle iteration. Run loops until its context is done.
	res, err := p.Poll(ctx)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Printf("claimed %d, completed %d\n", res.Claimed, res.Completed)

	// Output:
	// handling order 1
	// handling order 2
	// claimed 2, completed 2
}

func ExampleAny() {
	notifier := poller.NewNotifier()

	// Poll as soon as a writer notifies, and at least every 30 seconds for retries and entries written elsewhere.
	trigger := poller.Any(notifier, poller.Ticker(30*time.Second))

	notifier.Notify()

	err := trigger.Wait(context.Background(), poller.Result{})
	fmt.Println(err)

	// Output:
	// <nil>
}
