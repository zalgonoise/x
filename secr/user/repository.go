package user

import "context"

// Repository describes the actions exposed by the users store
type Repository interface {
	// Create will create a user `u`, returning its ID and an error
	Create(ctx context.Context, u *User) (uint64, error)
	// Get returns the user identified by `username`, and an error
	Get(ctx context.Context, username string) (*User, error)
	// List returns all the users, and an error
	List(ctx context.Context) ([]*User, error)
	// Update will update the user `username` with its updated version `updated`. Returns an error
	Update(ctx context.Context, username string, updated *User) error
	// Delete removes the user identified by `username`, returning an error
	Delete(ctx context.Context, username string) error
}
