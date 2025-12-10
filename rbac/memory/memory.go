package memory

import (
	"context"
	"slices"
	"sync"
)

type Memory[User comparable, Role comparable, Permission comparable] struct {
	rolePermissions map[Role][]Permission
	muRoles         *sync.RWMutex

	userRoles map[User][]Role
	muUsers   *sync.RWMutex
}

func NewMemory[User comparable, Role comparable, Permission comparable]() *Memory[User, Role, Permission] {
	return &Memory[User, Role, Permission]{
		rolePermissions: make(map[Role][]Permission),
		muRoles:         &sync.RWMutex{},
		userRoles:       make(map[User][]Role),
		muUsers:         &sync.RWMutex{},
	}
}

func (r *Memory[User, Role, Permission]) GrantPermission(_ context.Context, role Role, permission Permission) error {
	r.muRoles.Lock()

	if r.rolePermissions == nil {
		r.rolePermissions = make(map[Role][]Permission)
	}

	if r.rolePermissions[role] == nil {
		r.rolePermissions[role] = make([]Permission, 0, 64)
	}

	r.rolePermissions[role] = append(r.rolePermissions[role], permission)
	r.muRoles.Unlock()

	return nil
}

func (r *Memory[User, Role, Permission]) RevokePermission(_ context.Context, role Role, permission Permission) error {
	r.muRoles.Lock()
	defer r.muRoles.Unlock()

	if r.rolePermissions == nil {
		r.rolePermissions = make(map[Role][]Permission)

		return nil
	}

	if r.rolePermissions[role] == nil {
		return nil
	}

	permissions := r.rolePermissions[role]

	switch idx := slices.Index(permissions, permission); idx {
	case -1:
		return nil
	default:
		r.rolePermissions[role] = append(permissions[:idx], permissions[idx+1:]...)
	}

	return nil
}

func (r *Memory[User, Role, Permission]) GrantRole(_ context.Context, user User, role Role) error {
	r.muUsers.Lock()

	if r.userRoles == nil {
		r.userRoles = make(map[User][]Role)
	}

	if r.userRoles[user] == nil {
		r.userRoles[user] = make([]Role, 0, 64)
	}

	r.userRoles[user] = append(r.userRoles[user], role)
	r.muUsers.Unlock()

	return nil

}

func (r *Memory[User, Role, Permission]) RevokeRole(_ context.Context, user User, role Role) error {
	r.muUsers.Lock()
	defer r.muUsers.Unlock()

	if r.userRoles == nil {
		r.userRoles = make(map[User][]Role)

		return nil
	}

	if r.userRoles[user] == nil {
		return nil
	}

	roles := r.userRoles[user]

	switch idx := slices.Index(roles, role); idx {
	case -1:
		return nil
	default:
		r.userRoles[user] = append(roles[:idx], roles[idx+1:]...)
	}

	return nil

}

func (r *Memory[User, Role, Permission]) Can(_ context.Context, user User, permission Permission) bool {
	r.muUsers.RLock()

	roles, hasRoles := r.userRoles[user]
	if !hasRoles {
		r.muUsers.RUnlock()

		return false
	}

	r.muUsers.RUnlock()

	r.muRoles.RLock()
	for _, role := range roles {
		if slices.Contains(r.rolePermissions[role], permission) {
			r.muRoles.RUnlock()

			return true
		}
	}

	r.muRoles.RUnlock()

	return false
}
