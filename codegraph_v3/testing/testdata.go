package main

import (
	"context"
	"fmt"
	"os"
)

func Run(ctx context.Context, n int) error {
	for i := 0; i < n; i++ {
		_, err := Poll(i)
		if err != nil {
			return fmt.Errorf("failed to poll %d with err: %w", i, err)
		}
	}
	return nil
}

func Poll(n int) (int, error) {
	sqr := n * n
	if sqr%5 == 0 {
		Action(sqr)
	}

	return sqr, nil
}

func Action(n int) {
	fmt.Printf("%d is divisible by 5\n", n)
}

func Context(context.Context) {}

func Int(int) {}

func main() {
	err := Run(context.Background(), 10)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
