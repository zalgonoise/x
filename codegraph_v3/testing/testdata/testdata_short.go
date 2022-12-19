package testing

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

func main() {
	err := Run(context.Background(), 10)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
