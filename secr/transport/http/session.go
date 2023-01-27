package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/zalgonoise/x/ghttp"
	"github.com/zalgonoise/x/secr/authz"
	"github.com/zalgonoise/x/secr/service"
	"github.com/zalgonoise/x/secr/sqlite"
	"github.com/zalgonoise/x/secr/user"
)

func (s *server) login() http.HandlerFunc {
	type loginRequest struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}

	var parseFn = func(ctx context.Context, r *http.Request) (*loginRequest, error) {
		return ghttp.ReadBody[loginRequest](ctx, r)
	}

	var execFn = func(ctx context.Context, q *loginRequest) *ghttp.Response[user.Session] {
		if q == nil {
			return ghttp.NewResponse[user.Session](http.StatusBadRequest, "invalid request")
		}

		dbsession, err := s.s.Login(ctx, q.Username, q.Password)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.Session](http.StatusNotFound, err.Error())
			}
			if errors.Is(service.ErrIncorrectPassword, err) {
				return ghttp.NewResponse[user.Session](http.StatusBadRequest, err.Error())
			}
			return ghttp.NewResponse[user.Session](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.Session](http.StatusOK, "user logged in successfully").WithData(dbsession)
	}

	return ghttp.Do("Login", parseFn, execFn)
}

func (s *server) logout() http.HandlerFunc {
	var parseFn = func(ctx context.Context, r *http.Request) (*string, error) {
		caller, ok := authz.GetCaller(r)
		if ok {
			return &caller, nil
		}
		return nil, errors.New("failed to read password or token in request")
	}

	var execFn = func(ctx context.Context, q *string) *ghttp.Response[user.Session] {
		if q == nil || *q == "" {
			return ghttp.NewResponse[user.Session](http.StatusBadRequest, "invalid request")
		}

		err := s.s.Logout(ctx, *q)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.Session](http.StatusNotFound, err.Error())
			}
			return ghttp.NewResponse[user.Session](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.Session](http.StatusOK, "user logged out successfully")
	}

	return ghttp.Do("Logout", parseFn, execFn)
}

func (s *server) changePassword() http.HandlerFunc {
	type changePasswordRequest struct {
		Username    string `json:"-"`
		Password    string `json:"password,omitempty"`
		NewPassword string `json:"new_password,omitempty"`
	}

	var parseFn = func(ctx context.Context, r *http.Request) (*changePasswordRequest, error) {
		req, err := ghttp.ReadBody[changePasswordRequest](ctx, r)
		if err != nil {
			return nil, err
		}

		if caller, ok := authz.GetCaller(r); ok {
			req.Username = caller
			return req, nil
		}
		return nil, authz.ErrInvalidUser
	}

	var execFn = func(ctx context.Context, q *changePasswordRequest) *ghttp.Response[user.Session] {
		if q == nil {
			return ghttp.NewResponse[user.Session](http.StatusBadRequest, "invalid request")
		}

		err := s.s.ChangePassword(ctx, q.Username, q.Password, q.NewPassword)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.Session](http.StatusNotFound, err.Error())
			}
			if errors.Is(service.ErrIncorrectPassword, err) {
				return ghttp.NewResponse[user.Session](http.StatusBadRequest, err.Error())
			}
			return ghttp.NewResponse[user.Session](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.Session](http.StatusOK, "password changed successfully")
	}

	return ghttp.Do("ChangePassword", parseFn, execFn)
}

func (s *server) refresh() http.HandlerFunc {
	type refreshRequest struct {
		Username string `json:"username,omitempty"`
		Token    string `json:"-"`
	}
	var parseFn = func(ctx context.Context, r *http.Request) (*refreshRequest, error) {
		q, err := ghttp.ReadBody[refreshRequest](ctx, r)
		if err != nil {
			return nil, err
		}
		if token, ok := getToken(r); ok {
			q.Token = token
		}
		if caller, ok := authz.GetCaller(r); ok && caller == q.Username {
			return q, nil
		}

		return nil, authz.ErrInvalidUser
	}

	var execFn = func(ctx context.Context, q *refreshRequest) *ghttp.Response[user.Session] {
		if q == nil {
			return ghttp.NewResponse[user.Session](http.StatusBadRequest, "invalid request")
		}

		token, err := s.s.Refresh(ctx, q.Username, q.Token)
		if err != nil {
			if errors.Is(sqlite.ErrNotFoundUser, err) {
				return ghttp.NewResponse[user.Session](http.StatusNotFound, err.Error())
			}
			if errors.Is(service.ErrIncorrectPassword, err) {
				return ghttp.NewResponse[user.Session](http.StatusBadRequest, err.Error())
			}
			return ghttp.NewResponse[user.Session](http.StatusInternalServerError, err.Error())
		}

		return ghttp.NewResponse[user.Session](http.StatusOK, "token refreshed successfully").WithData(token)
	}

	return ghttp.Do("Refresh", parseFn, execFn)
}
