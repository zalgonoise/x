package endpoints

import (
	"net/http"

	"github.com/zalgonoise/x/secr/user"
)

func (e endpoints) Login(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.Login")
	defer s.End()

	u, err := readBody[user.User](ctx, r)
	if err != nil {
		res := NewResponse[user.User](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	uSession, err := e.s.Login(ctx, u.Username, u.Password)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.User](500, "login failed", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse(200, "logged in successfully", nil, uSession)
	res.WriteHTTP(ctx, w)
}

func (e endpoints) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.Logout")
	defer s.End()

	u, err := readBody[user.User](ctx, r)
	if err != nil {
		res := NewResponse[user.User](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	err = e.s.Logout(ctx, u.Username)
	if err != nil {
		res := NewResponse[user.User](500, "failed to logout", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse[user.User](200, "logged out successfully", nil, nil)
	res.WriteHTTP(ctx, w)
}

func (e endpoints) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.ChangePassword")
	defer s.End()

	npu, err := readBody[user.NewPassword](ctx, r)
	if err != nil {
		res := NewResponse[user.NewPassword](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	err = e.s.ChangePassword(ctx, npu.User.Username, npu.User.Password, npu.Password)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.NewPassword](500, "login failed", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse[user.User](200, "password changed successfully", nil, nil)
	res.WriteHTTP(ctx, w)
}

func (e endpoints) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, s := e.newCtxAndSpan(r, "http.Refresh")
	defer s.End()

	u, err := readBody[user.Session](ctx, r)
	if err != nil {
		res := NewResponse[user.Session](400, "failed to read record from body", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	uSession, err := e.s.Refresh(ctx, u.Username, u.Token)
	if err != nil {
		// TODO: check error type first
		res := NewResponse[user.Session](500, "session refresh failed", err, nil)
		res.WriteHTTP(ctx, w)
		return
	}

	res := NewResponse(200, "session refreshed successfully", nil, uSession)
	res.WriteHTTP(ctx, w)
}
