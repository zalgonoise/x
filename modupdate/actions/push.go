package actions

import "context"

func (a *ModUpdate) Push(ctx context.Context) error {
	// git add go.mod go.sum
	// git commit -m 'chore: updated modules'
	// git push

	return nil
}
