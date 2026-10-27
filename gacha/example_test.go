package gacha_test

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/zalgonoise/x/gacha"
)

func ExampleJSON() {
	type Order struct {
		ID    int    `json:"id"`
		Total string `json:"total"`
	}

	codec := gacha.JSON[Order]()

	data, err := codec.Marshal(Order{ID: 1, Total: "9.99"})
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(string(data))

	order, err := codec.Unmarshal(data)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(order.ID, order.Total)

	// Output:
	// {"id":1,"total":"9.99"}
	// 1 9.99
}

func ExamplePermanent() {
	errInvalidOrder := errors.New("invalid order")

	handle := func(_ context.Context, total int) error {
		if total < 0 {
			// Retrying will not help: move the entry to the dead-letter table right away.
			return gacha.Permanent(fmt.Errorf("total %d: %w", total, errInvalidOrder))
		}

		return nil
	}

	err := handle(context.Background(), -1)

	fmt.Println(err)
	fmt.Println(gacha.IsPermanent(err), errors.Is(err, errInvalidOrder))

	// Output:
	// total -1: invalid order
	// true true
}

func ExampleIDFromContext() {
	id := uuid.MustParse("019a0f5e-7c3a-7d2e-9b1f-3c4d5e6f7a8b")

	handle := func(ctx context.Context, _ string) error {
		key, ok := gacha.IDFromContext(ctx)
		if !ok {
			return errors.New("no entry ID")
		}

		// Use the entry ID as an idempotency key with downstream systems.
		fmt.Println("idempotency key:", key)

		return nil
	}

	_ = handle(gacha.ContextWithID(context.Background(), id), "item")

	// Output:
	// idempotency key: 019a0f5e-7c3a-7d2e-9b1f-3c4d5e6f7a8b
}
