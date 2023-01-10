package endpoints

import (
	"net/http"

	"github.com/zalgonoise/x/secr/user"
)

func (e endpoints) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.CreateUser")
	defer s.End()

	u, err := readBody[user.User](ctx, r)
	if err != nil {
		res := NewResponse[user.User](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	newUser, err := e.s.CreateUser(ctx, u.Username, u.Password, u.Name)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.User](500, "failed to create user", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse(200, "user created successfully", nil, newUser)
	res.WriteHTTP(ctx, w)
}

func (e endpoints) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.GetUser")
	defer s.End()

	u, err := readBody[user.User](ctx, r)
	if err != nil {
		res := NewResponse[user.User](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	dbuser, err := e.s.GetUser(ctx, u.Username)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.User](500, "failed to fetch user", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse(200, "user fetched successfully", nil, dbuser)
	res.WriteHTTP(ctx, w)
}

func (e endpoints) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.ListUsers")
	defer s.End()

	users, err := e.s.ListUsers(ctx)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[[]user.User](500, "failed to list users", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse(200, "users listed successfully", nil, &users)
	res.WriteHTTP(ctx, w)
}

func (e endpoints) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.UpdateUser")
	defer s.End()

	u, err := readBody[user.User](ctx, r)
	if err != nil {
		res := NewResponse[user.User](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	err = e.s.UpdateUser(ctx, u.Username, u)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.User](500, "failed to update user", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	newUser, err := e.s.GetUser(ctx, u.Username)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.User](500, "failed to fetch new user", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse(200, "user updated successfully", nil, newUser)
	res.WriteHTTP(ctx, w)
}

func (e endpoints) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.DeleteUser")
	defer s.End()

	u, err := readBody[user.User](ctx, r)
	if err != nil {
		res := NewResponse[user.User](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	err = e.s.DeleteUser(ctx, u.Username)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.User](500, "failed to delete user", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse[*user.User](200, "user deleted successfully", nil, nil)
	res.WriteHTTP(ctx, w)
}
