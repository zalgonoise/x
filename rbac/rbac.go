package rbac

import "context"

type AccessSystem[User comparable, Role comparable, Permission comparable] interface {
	GrantPermission(ctx context.Context, role Role, permission Permission) error
	RevokePermission(ctx context.Context, role Role, permission Permission) error
	GrantRole(ctx context.Context, user User, role Role) error
	RevokeRole(ctx context.Context, user User, role Role) error

	Can(ctx context.Context, user User, permission Permission) bool
}
