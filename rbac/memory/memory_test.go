package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemory_GrantPermission(t *testing.T) {
	for _, testcase := range []struct {
		name   string
		input  map[string][]string
		wants  map[string][]string
		prepFn func(m *Memory[string, string, string])
		err    error
	}{
		{
			name:  "Success/RoleWithPermissions",
			input: map[string][]string{"users.admin": {"read", "create", "update", "delete"}},
			wants: map[string][]string{"users.admin": {"read", "create", "update", "delete"}},
		},
		{
			name: "Success/MultipleRolesWithPermissions",
			input: map[string][]string{
				"users.admin": {"read", "create", "update", "delete"},
				"users.audit": {"read"},
			},
			wants: map[string][]string{
				"users.admin": {"read", "create", "update", "delete"},
				"users.audit": {"read"},
			},
		},
		{
			name:  "Success/NoRoles",
			input: map[string][]string{},
			wants: map[string][]string{},
		},
		{
			name:  "Success/NilRoles",
			input: nil,
			wants: map[string][]string{},
		},
		{
			name: "Success/NilRolesInDataStructure",
			prepFn: func(m *Memory[string, string, string]) {
				m.muRoles.Lock()
				m.rolePermissions = nil
				m.muRoles.Unlock()
			},
			input: map[string][]string{"users.admin": {"read", "create", "update", "delete"}},
			wants: map[string][]string{"users.admin": {"read", "create", "update", "delete"}},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			m := NewMemory[string, string, string]()
			ctx := context.Background()

			if testcase.prepFn != nil {
				testcase.prepFn(m)
			}

			for role, permissions := range testcase.input {
				for i := range permissions {
					if err := m.GrantPermission(ctx, role, permissions[i]); err != nil {
						require.ErrorIs(t, err, testcase.err)

						return
					}
				}
			}

			res := m.rolePermissions

			require.Len(t, res, len(testcase.wants))
			for role, permissions := range testcase.wants {
				resPermissions, ok := res[role]

				require.True(t, ok)
				require.Len(t, resPermissions, len(permissions))
				require.Equal(t, permissions, resPermissions)
			}
		})
	}
}

func TestMemory_RevokePermission(t *testing.T) {
	for _, testcase := range []struct {
		name        string
		input       map[string][]string
		revocations map[string][]string
		wants       map[string][]string
		prepFn      func(m *Memory[string, string, string])
		err         error
	}{
		{
			name:        "Success/RoleWithPermissions/NoSuitableRevocations",
			input:       map[string][]string{"users.admin": {"read", "create", "update", "delete"}},
			revocations: map[string][]string{"users.audit": {"read"}},
			wants:       map[string][]string{"users.admin": {"read", "create", "update", "delete"}},
		},
		{
			name: "Success/MultipleRolesWithPermissions/SuitableRevocations",
			input: map[string][]string{
				"users.admin": {"read", "create", "update", "delete"},
				"users.audit": {"read"},
			},
			revocations: map[string][]string{"users.audit": {"read"}},
			wants: map[string][]string{
				"users.admin": {"read", "create", "update", "delete"},
				"users.audit": {},
			},
		},
		{
			name:        "Success/RoleWithPermissions/PermissionUnassigned",
			input:       map[string][]string{"users.admin": {"read", "create", "update", "delete"}},
			revocations: map[string][]string{"users.admin": {"patch"}},
			wants:       map[string][]string{"users.admin": {"read", "create", "update", "delete"}},
		},
		{
			name:  "Success/NoRoles",
			input: map[string][]string{},
			wants: map[string][]string{},
		},
		{
			name:  "Success/NilRoles",
			input: nil,
			wants: map[string][]string{},
		},
		{
			name: "Success/NilRolesInDataStructure",
			prepFn: func(m *Memory[string, string, string]) {
				m.muRoles.Lock()
				m.rolePermissions = nil
				m.muRoles.Unlock()
			},
			input:       map[string][]string{"users.admin": {"read", "create", "update", "delete"}},
			revocations: map[string][]string{"users.audit": {"read"}},
			wants:       map[string][]string{},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			m := NewMemory[string, string, string]()
			ctx := context.Background()

			for role, permissions := range testcase.input {
				for i := range permissions {
					if err := m.GrantPermission(ctx, role, permissions[i]); err != nil {
						require.ErrorIs(t, err, testcase.err)

						return
					}
				}
			}

			if testcase.prepFn != nil {
				testcase.prepFn(m)
			}

			for role, permissions := range testcase.revocations {
				for i := range permissions {
					if err := m.RevokePermission(ctx, role, permissions[i]); err != nil {
						require.ErrorIs(t, err, testcase.err)

						return
					}
				}
			}

			res := m.rolePermissions

			require.Len(t, res, len(testcase.wants))
			for role, permissions := range testcase.wants {
				resPermissions, ok := res[role]

				require.True(t, ok)
				require.Len(t, resPermissions, len(permissions))
				require.Equal(t, permissions, resPermissions)
			}
		})
	}
}

func TestMemory_GrantRole(t *testing.T) {
	for _, testcase := range []struct {
		name   string
		input  map[string][]string
		wants  map[string][]string
		prepFn func(m *Memory[string, string, string])
		err    error
	}{
		{
			name:  "Success/UserWithRoles",
			input: map[string][]string{"gopher": {"users.admin", "users.audit"}},
			wants: map[string][]string{"gopher": {"users.admin", "users.audit"}},
		},
		{
			name: "Success/MultipleUsersWithRoles",
			input: map[string][]string{
				"gopher":  {"users.admin", "users.audit"},
				"auditor": {"users.audit"},
			},
			wants: map[string][]string{
				"gopher":  {"users.admin", "users.audit"},
				"auditor": {"users.audit"},
			},
		},
		{
			name:  "Success/NoUsers",
			input: map[string][]string{},
			wants: map[string][]string{},
		},
		{
			name:  "Success/NilUsers",
			input: nil,
			wants: map[string][]string{},
		},
		{
			name: "Success/NilUsersInDataStructure",
			prepFn: func(m *Memory[string, string, string]) {
				m.muUsers.Lock()
				m.userRoles = nil
				m.muUsers.Unlock()
			},
			input: map[string][]string{"gopher": {"users.admin", "users.audit"}},
			wants: map[string][]string{"gopher": {"users.admin", "users.audit"}},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			m := NewMemory[string, string, string]()
			ctx := context.Background()

			if testcase.prepFn != nil {
				testcase.prepFn(m)
			}

			for user, roles := range testcase.input {
				for i := range roles {
					if err := m.GrantRole(ctx, user, roles[i]); err != nil {
						require.ErrorIs(t, err, testcase.err)

						return
					}
				}
			}

			res := m.userRoles

			require.Len(t, res, len(testcase.wants))
			for user, roles := range testcase.wants {
				resRoles, ok := res[user]

				require.True(t, ok)
				require.Len(t, resRoles, len(roles))
				require.Equal(t, roles, resRoles)
			}
		})
	}
}

func TestMemory_RevokeRole(t *testing.T) {
	for _, testcase := range []struct {
		name        string
		input       map[string][]string
		revocations map[string][]string
		wants       map[string][]string
		prepFn      func(m *Memory[string, string, string])
		err         error
	}{
		{
			name:        "Success/UserWithRoles/NoSuitableRevocations",
			input:       map[string][]string{"gopher": {"users.admin", "users.audit"}},
			revocations: map[string][]string{"auditor": {"users.audit"}},
			wants:       map[string][]string{"gopher": {"users.admin", "users.audit"}},
		},
		{
			name: "Success/MultipleUsersWithRoles/SuitableRevocations",
			input: map[string][]string{
				"gopher":  {"users.admin", "users.audit"},
				"auditor": {"users.audit"},
			},
			revocations: map[string][]string{"auditor": {"users.audit"}},
			wants: map[string][]string{
				"gopher":  {"users.admin", "users.audit"},
				"auditor": {},
			},
		},
		{
			name:        "Success/UserWithRoles/PermissionUnassigned",
			input:       map[string][]string{"gopher": {"users.admin", "users.audit"}},
			revocations: map[string][]string{"gopher": {"groups.admin"}},
			wants:       map[string][]string{"gopher": {"users.admin", "users.audit"}},
		},
		{
			name:  "Success/NoRoles",
			input: map[string][]string{},
			wants: map[string][]string{},
		},
		{
			name:  "Success/NilRoles",
			input: nil,
			wants: map[string][]string{},
		},
		{
			name: "Success/NilUsersInDataStructure",
			prepFn: func(m *Memory[string, string, string]) {
				m.muUsers.Lock()
				m.userRoles = nil
				m.muUsers.Unlock()
			},
			input:       map[string][]string{"gopher": {"users.admin", "users.audit"}},
			revocations: map[string][]string{"gopher": {"users.audit"}},
			wants:       map[string][]string{},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			m := NewMemory[string, string, string]()
			ctx := context.Background()

			for user, roles := range testcase.input {
				for i := range roles {
					if err := m.GrantRole(ctx, user, roles[i]); err != nil {
						require.ErrorIs(t, err, testcase.err)

						return
					}
				}
			}

			if testcase.prepFn != nil {
				testcase.prepFn(m)
			}

			for user, roles := range testcase.revocations {
				for i := range roles {
					if err := m.RevokeRole(ctx, user, roles[i]); err != nil {
						require.ErrorIs(t, err, testcase.err)

						return
					}
				}
			}

			res := m.userRoles

			require.Len(t, res, len(testcase.wants))
			for user, roles := range testcase.wants {
				resRoles, ok := res[user]

				require.True(t, ok)
				require.Len(t, resRoles, len(roles))
				require.Equal(t, roles, resRoles)
			}
		})
	}
}

func TestMemory_Can(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		inputRoles map[string][]string
		inputUsers map[string][]string
		user       string
		role       string
		wants      bool
		err        error
	}{
		{
			name: "Success/UserWithRoles/HasRole",
			inputRoles: map[string][]string{
				"users.admin": {"read", "create", "update", "delete"},
				"users.audit": {"read"},
			},
			inputUsers: map[string][]string{"gopher": {"users.admin", "users.audit"}},
			user:       "gopher",
			role:       "read",
			wants:      true,
		},
		{
			name: "Success/UserWithRoles/DoesNotHaveRole",
			inputRoles: map[string][]string{
				"users.admin": {"read", "create", "update", "delete"},
				"users.audit": {"read"},
			},
			inputUsers: map[string][]string{"gopher": {"users.admin", "users.audit"}},
			user:       "gopher",
			role:       "patch",
			wants:      false,
		},
		{
			name: "Success/UserDoesNotExist",
			inputRoles: map[string][]string{
				"users.admin": {"read", "create", "update", "delete"},
				"users.audit": {"read"},
			},
			inputUsers: map[string][]string{"gopher": {"users.admin", "users.audit"}},
			user:       "auditor",
			role:       "read",
			wants:      false,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			m := NewMemory[string, string, string]()
			ctx := context.Background()

			for role, permissions := range testcase.inputRoles {
				for i := range permissions {
					if err := m.GrantPermission(ctx, role, permissions[i]); err != nil {
						require.ErrorIs(t, err, testcase.err)

						return
					}
				}
			}

			for user, roles := range testcase.inputUsers {
				for i := range roles {
					if err := m.GrantRole(ctx, user, roles[i]); err != nil {
						require.ErrorIs(t, err, testcase.err)

						return
					}
				}
			}

			require.Equal(t, testcase.wants, m.Can(ctx, testcase.user, testcase.role))
		})
	}
}
